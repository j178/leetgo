use anyhow::Result;

pub fn read_line() -> Result<String> {
    let mut line = String::new();
    std::io::stdin().read_line(&mut line)?;
    Ok(line)
}
