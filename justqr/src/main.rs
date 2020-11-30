use actix_files::Files;
use actix_web::{get, middleware, App, HttpResponse, HttpServer, Result};
use qrcode::render::svg;
use qrcode::QrCode;

#[get("/api/code")]
async fn get_code() -> Result<HttpResponse> {
    let svg = QrCode::new(b"https://app.zenhub.com/workspaces/toolsmiths-5e57d39702cc3e041fa62d72/issues/wearethorn/platform/766")
        .unwrap()
        .render::<svg::Color>()
        .build();
    Ok(HttpResponse::Ok().content_type("image/svg+xml").body(svg))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .wrap(middleware::Logger::default())
            .service(get_code)
            .service(Files::new("/", "./static").index_file("index.html"))
    })
    .bind("127.0.0.1:8080")?
    .run()
    .await
}
