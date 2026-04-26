fn example() -> Result<(), Box<dyn std::error::Error>> {
    let data = std::fs::read_to_string("file.txt")?;
    Ok(())
}