use actix_files::Files;
use actix_web::{middleware, post, web, App, HttpResponse, HttpServer, Result};
use qrcode::render::svg;
use qrcode::QrCode;
use serde::Deserialize;

#[derive(Deserialize)]
struct CodeRequest {
    content: String,
}

fn code_to_svg(code: &[u8]) -> String {
    QrCode::new(code)
        .unwrap()
        .render::<svg::Color>()
        .quiet_zone(false)
        .max_dimensions(300, 300)
        .build()
}

#[post("/api/svg")]
async fn create_svg(cr: web::Form<CodeRequest>) -> Result<HttpResponse> {
    let svg = code_to_svg(cr.content.as_bytes());
    Ok(HttpResponse::Ok().content_type("image/svg+xml").body(svg))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .wrap(middleware::Logger::default())
            .service(create_svg)
            .service(Files::new("/", "./static").index_file("index.html"))
    })
    .bind("0.0.0.0:8080")?
    .run()
    .await
}
