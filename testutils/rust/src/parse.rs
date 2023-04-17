use anyhow::{bail, Result};
use serde::{Deserialize, Serialize};
use serde_json::Value;

pub fn deserialize<'de, T: Deserialize<'de>>(s: &'de str) -> Result<T> {
    let res: T = serde_json::from_str(s)?;
    Ok(res)
}

pub fn serialize<T: Serialize>(v: T) -> Result<String> {
    let res = serde_json::to_string(&v)?;
    Ok(res)
}

pub fn split_array(raw: &str) -> Result<Vec<String>> {
    let trimmed = raw.trim();

    if trimmed.len() <= 1 || !trimmed.starts_with('[') || !trimmed.ends_with(']') {
        bail!("invalid array: {}", trimmed);
    }

    let splits: Vec<Value> = serde_json::from_str(trimmed)?;
    let res: Vec<String> = splits.iter().map(|v| v.to_string()).collect();
    Ok(res)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_split_array() {
        let test_cases = vec![
            ("[]", vec![]),
            ("[1]", vec!["1"]),
            (r#"["a", "b"]"#, vec![r#""a""#, r#""b""#]),
            ("[1, 2, 3]", vec!["1", "2", "3"]),
            (r#"[1, "a", null, true, false]"#, vec!["1", r#""a""#, "null", "true", "false"]),
            ("[1, [2, 3], 4]", vec!["1", "[2,3]", "4"]),
            ("   [1, 2]  ", vec!["1", "2"]),
        ];

        for (input, expected) in test_cases {
            let result = split_array(input);
            match result {
                Ok(res) => assert_eq!(res, expected),
                Err(_) => panic!("Test failed for input: {}", input),
            }
        }
    }
}
