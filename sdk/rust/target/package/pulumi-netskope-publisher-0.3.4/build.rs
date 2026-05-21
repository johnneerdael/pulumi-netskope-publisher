use std::error::Error;
use std::path::PathBuf;

fn main() -> Result<(), Box<dyn Error>> {
    let schema = PathBuf::from(std::env::var("CARGO_MANIFEST_DIR")?).join("schema.json");
    pulumi_gestalt_build::generate_from_schema(&schema)?;
    Ok(())
}
