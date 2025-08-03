//! main.rs
#[macro_use]
extern crate rocket;

use rocket::fairing::AdHoc;
use sqlx::PgPool;

mod config_db;
mod model;
mod routes;

#[launch]
fn rocket() -> _ {
    rocket::build()
        .attach(AdHoc::on_ignite("Database Config", config_db::init_pool))
        .mount("/", routes::routes())
}
