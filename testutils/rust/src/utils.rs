use anyhow::Result;

pub fn read_line() -> Result<String> {
    let mut line = String::new();
    std::io::stdin().read_line(&mut line)?;
    Ok(line)
}

pub fn join_array(arr: Vec<String>) -> String {
    "[".to_string() + &arr.join(",") + "]"
}
