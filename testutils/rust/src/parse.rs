use anyhow::Result;
use serde::{Deserialize, Serialize};

pub fn deserialize<'de, T: Deserialize<'de>>(s: &'de str) -> Result<T> {
    let res: T = serde_json::from_str(s)?;
    Ok(res)
}

pub fn serialize<T: Serialize>(v: T) -> Result<String> {
    let res = serde_json::to_string(&v)?;
    Ok(res)
}

