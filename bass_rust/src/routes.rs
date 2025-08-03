use crate::config_db::ConfigDb;
use crate::model::{ApiResponse, ErrorResponse};
use rocket::{
    Route, State, delete, get, http::Status, patch, post, put, response::status, routes,
    serde::json::Json,
};
use serde_json::{Value, json};
use sqlx::{Column, PgPool, Row, postgres::PgPoolOptions};

pub fn routes() -> Vec<Route> {
    routes![handle_dynamic_api, health_check]
}

#[get("/health")]
pub async fn health_check() -> Json<Value> {
    Json(json!({
        "status": "healthy",
        "service": "BaaS Dynamic API Handler"
    }))
}

// Dynamic route handler that matches: /api/{project_id}{api_path}
#[get("/api/<project_id>/<path..>")]
pub async fn handle_dynamic_api(
    config_db: &State<ConfigDb>,
    project_id: &str,
    path: std::path::PathBuf,
) -> Result<Json<ApiResponse>, status::Custom<Json<ErrorResponse>>> {
    handle_request(
        config_db,
        project_id,
        &format!("/{}", path.display()),
        "GET",
    )
    .await
}

#[post("/api/<project_id>/<path..>", data = "<body>")]
pub async fn handle_dynamic_post(
    config_db: &State<ConfigDb>,
    project_id: &str,
    path: std::path::PathBuf,
    body: Json<Value>,
) -> Result<Json<ApiResponse>, status::Custom<Json<ErrorResponse>>> {
    handle_request_with_body(
        config_db,
        project_id,
        &format!("/{}", path.display()),
        "POST",
        Some(body.into_inner()),
    )
    .await
}

#[put("/api/<project_id>/<path..>", data = "<body>")]
pub async fn handle_dynamic_put(
    config_db: &State<ConfigDb>,
    project_id: &str,
    path: std::path::PathBuf,
    body: Json<Value>,
) -> Result<Json<ApiResponse>, status::Custom<Json<ErrorResponse>>> {
    handle_request_with_body(
        config_db,
        project_id,
        &format!("/{}", path.display()),
        "PUT",
        Some(body.into_inner()),
    )
    .await
}

#[delete("/api/<project_id>/<path..>")]
pub async fn handle_dynamic_delete(
    config_db: &State<ConfigDb>,
    project_id: &str,
    path: std::path::PathBuf,
) -> Result<Json<ApiResponse>, status::Custom<Json<ErrorResponse>>> {
    handle_request(
        config_db,
        project_id,
        &format!("/{}", path.display()),
        "DELETE",
    )
    .await
}

async fn handle_request(
    config_db: &State<ConfigDb>,
    project_id: &str,
    api_path: &str,
    method: &str,
) -> Result<Json<ApiResponse>, status::Custom<Json<ErrorResponse>>> {
    handle_request_with_body(config_db, project_id, api_path, method, None).await
}

async fn handle_request_with_body(
    config_db: &State<ConfigDb>,
    project_id: &str,
    api_path: &str,
    method: &str,
    body: Option<Value>,
) -> Result<Json<ApiResponse>, status::Custom<Json<ErrorResponse>>> {
    // Get project and APIs from config database
    let (project, apis) = config_db
        .get_project_with_apis(project_id)
        .await
        .map_err(|e| {
            status::Custom(
                Status::NotFound,
                Json(ErrorResponse {
                    error: "Project not found".to_string(),
                    message: format!("Project with ID '{}' not found: {}", project_id, e),
                }),
            )
        })?;

    // Find matching API configuration
    let matching_api = apis
        .iter()
        .find(|api| api.path == api_path && api.method.to_uppercase() == method.to_uppercase())
        .ok_or_else(|| {
            status::Custom(
                Status::Forbidden,
                Json(ErrorResponse {
                    error: "API not configured".to_string(),
                    message: format!(
                        "No {} API configured for path '{}' in project '{}'",
                        method, api_path, project_id
                    ),
                }),
            )
        })?;

    // Connect to user's database
    let user_db = PgPoolOptions::new()
        .max_connections(5)
        .connect(&project.database_url)
        .await
        .map_err(|e| {
            status::Custom(
                Status::InternalServerError,
                Json(ErrorResponse {
                    error: "Database connection failed".to_string(),
                    message: format!("Failed to connect to project database: {}", e),
                }),
            )
        })?;

    // Execute the appropriate database operation
    let result = match method.to_uppercase().as_str() {
        "GET" => handle_get_request(&user_db, &matching_api.table_name).await,
        "POST" => handle_post_request(&user_db, &matching_api.table_name, body).await,
        "PUT" => handle_put_request(&user_db, &matching_api.table_name, body).await,
        "DELETE" => handle_delete_request(&user_db, &matching_api.table_name).await,
        _ => Err(format!("Unsupported HTTP method: {}", method)),
    }
    .map_err(|e| {
        status::Custom(
            Status::InternalServerError,
            Json(ErrorResponse {
                error: "Database operation failed".to_string(),
                message: e,
            }),
        )
    })?;

    Ok(Json(result))
}

async fn handle_get_request(pool: &PgPool, table_name: &str) -> Result<ApiResponse, String> {
    let query = format!("SELECT * FROM {}", table_name);
    let rows = sqlx::query(&query)
        .fetch_all(pool)
        .await
        .map_err(|e| format!("Query execution failed: {}", e))?;

    let mut result = Vec::new();
    for row in rows {
        let mut map = serde_json::Map::new();
        for (i, col) in row.columns().iter().enumerate() {
            let name = col.name();

            // Try different types in order of likelihood
            let val: Value = if let Ok(v) = row.try_get::<i32, _>(i) {
                json!(v)
            } else if let Ok(v) = row.try_get::<i64, _>(i) {
                json!(v)
            } else if let Ok(v) = row.try_get::<String, _>(i) {
                json!(v)
            } else if let Ok(v) = row.try_get::<bool, _>(i) {
                json!(v)
            } else if let Ok(v) = row.try_get::<chrono::DateTime<chrono::Utc>, _>(i) {
                json!(v)
            } else if let Ok(v) = row.try_get::<f64, _>(i) {
                json!(v)
            } else {
                Value::Null
            };

            map.insert(name.to_string(), val);
        }
        result.push(Value::Object(map));
    }

    Ok(ApiResponse {
        data: json!({ "items": result }),
        count: Some(result.len()),
    })
}

async fn handle_post_request(
    pool: &PgPool,
    table_name: &str,
    body: Option<Value>,
) -> Result<ApiResponse, String> {
    let body = body.ok_or("Request body is required for POST requests")?;

    if let Value::Object(map) = body {
        let columns: Vec<String> = map.keys().cloned().collect();
        let placeholders: Vec<String> = (1..=columns.len()).map(|i| format!("${}", i)).collect();

        let query = format!(
            "INSERT INTO {} ({}) VALUES ({}) RETURNING *",
            table_name,
            columns.join(", "),
            placeholders.join(", ")
        );

        let mut query_builder = sqlx::query(&query);
        for col in &columns {
            query_builder = query_builder.bind(map.get(col).unwrap_or(&Value::Null));
        }

        let row = query_builder
            .fetch_one(pool)
            .await
            .map_err(|e| format!("Insert failed: {}", e))?;

        let mut result_map = serde_json::Map::new();
        for col in row.columns() {
            let name = col.name();
            let val: Value = row
                .try_get::<serde_json::Value, _>(name)
                .unwrap_or(Value::Null);
            result_map.insert(name.to_string(), val);
        }

        Ok(ApiResponse {
            data: json!({ "created": Value::Object(result_map) }),
            count: Some(1),
        })
    } else {
        Err("Request body must be a JSON object".to_string())
    }
}

async fn handle_put_request(
    pool: &PgPool,
    table_name: &str,
    body: Option<Value>,
) -> Result<ApiResponse, String> {
    let body = body.ok_or("Request body is required for PUT requests")?;

    if let Value::Object(map) = body {
        // For PUT, we assume there's an 'id' field for the WHERE clause
        let id = map.get("id").ok_or("PUT request requires an 'id' field")?;

        let mut updates = Vec::new();
        let mut values = Vec::new();
        let mut param_count = 1;

        for (key, value) in &map {
            if key != "id" {
                updates.push(format!("{} = ${}", key, param_count));
                values.push(value);
                param_count += 1;
            }
        }

        let query = format!(
            "UPDATE {} SET {} WHERE id = ${} RETURNING *",
            table_name,
            updates.join(", "),
            param_count
        );

        let mut query_builder = sqlx::query(&query);
        for value in values {
            query_builder = query_builder.bind(value);
        }
        query_builder = query_builder.bind(id);

        let row = query_builder
            .fetch_one(pool)
            .await
            .map_err(|e| format!("Update failed: {}", e))?;

        let mut result_map = serde_json::Map::new();
        for col in row.columns() {
            let name = col.name();
            let val: Value = row
                .try_get::<serde_json::Value, _>(name)
                .unwrap_or(Value::Null);
            result_map.insert(name.to_string(), val);
        }

        Ok(ApiResponse {
            data: json!({ "updated": Value::Object(result_map) }),
            count: Some(1),
        })
    } else {
        Err("Request body must be a JSON object".to_string())
    }
}

async fn handle_delete_request(pool: &PgPool, table_name: &str) -> Result<ApiResponse, String> {
    // For now, we'll implement a simple DELETE ALL - you might want to add WHERE conditions
    let query = format!("DELETE FROM {}", table_name);
    let result = sqlx::query(&query)
        .execute(pool)
        .await
        .map_err(|e| format!("Delete failed: {}", e))?;

    Ok(ApiResponse {
        data: json!({ "message": "Records deleted successfully" }),
        count: Some(result.rows_affected() as usize),
    })
}
