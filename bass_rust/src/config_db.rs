use rocket::{Build, Rocket, State};
use sqlx::{PgPool, postgres::PgPoolOptions};
use std::env;

pub struct ConfigDb(pub PgPool);

impl ConfigDb {
    pub async fn get_project_with_apis(
        &self,
        project_id: &str,
    ) -> Result<(crate::model::Project, Vec<crate::model::API>), sqlx::Error> {
        let project = sqlx::query_as::<_, crate::model::Project>(
            "SELECT * FROM projects WHERE id = $1 AND status = 'active'",
        )
        .bind(project_id)
        .fetch_one(&self.0)
        .await?;

        let apis = sqlx::query_as::<_, crate::model::API>(
            "SELECT * FROM apis WHERE project_id = $1 AND status = 'active'",
        )
        .bind(project_id)
        .fetch_all(&self.0)
        .await?;

        Ok((project, apis))
    }
}

pub async fn init_pool(rocket: Rocket<Build>) -> Rocket<Build> {
    let database_url = env::var("DATABASE_URL")
        .unwrap_or_else(|_| "postgres://baas_user:8891520703@localhost:5432/backend_automation_service?sslmode=disable".to_string());

    match PgPoolOptions::new()
        .max_connections(10)
        .connect(&database_url)
        .await
    {
        Ok(pool) => rocket.manage(ConfigDb(pool)),
        Err(e) => {
            eprintln!("Failed to connect to database: {}", e);
            std::process::exit(1);
        }
    }
}
