use lambda::handler_fn;
use log::info;
use serde_json::Value;
use std::env;

type Error = Box<dyn std::error::Error + Send + Sync + 'static>;

#[tokio::main]
async fn main() -> Result<(), Error> {
    simple_logger::init_with_level(log::Level::Debug)?;
    lambda::run(handler_fn(handler)).await?;
    Ok(())
}

// Return the input event as output
async fn handler(event: Value) -> Result<Value, Error> {
    info!("Hello from {}!", env::var("AWS_LAMBDA_FUNCTION_NAME")?);
    Ok(event)
}
