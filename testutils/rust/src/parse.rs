use std::error::Error;
use std::str::FromStr;

use serde::Deserialize;
use serde_json::Value;

use crate::{ListNode, TreeNode};

pub fn split_array(raw: &str) -> Result<Vec<String>, Box<dyn Error>> {
    let trimmed = raw.trim();

    if trimmed.len() <= 1 || !trimmed.starts_with('[') || !trimmed.ends_with(']') {
        return Err(format!("Invalid array: {}", trimmed).into());
    }

    let splits: Vec<Value> = serde_json::from_str(trimmed)?;
    let res: Vec<String> = splits.iter().map(|v| v.to_string()).collect();
    Ok(res)
}

#[derive(Debug, PartialEq)]
pub struct Array<T>(pub Vec<T>);

impl<'de, T: Deserialize<'de>> Deserialize<'de> for Array<T> {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
        where
            D: serde::de::Deserializer<'de>,
    {
        let splits = split_array(&String::deserialize(deserializer)?).unwrap();
        let res: Result<Vec<T>, D::Error> = splits.iter().map(|s| T::deserialize(s)).collect();
        Ok(Array(res.unwrap()))
    }
}

pub fn parse<'de, T: Deserialize<'de>>(s: &str) -> Result<T, Box<dyn Error>> {
    let res: T = serde_json::from_str(s)?;
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

    #[test]
    fn test_parse_basic() {
        let result = parse::<i32>("1");
        assert_eq!(result.unwrap(), 1);
        let result = parse::<Array<i32>>("[1, 2, 3]");
        assert_eq!(result.unwrap(), Array(vec![1, 2, 3]));
    }
}
