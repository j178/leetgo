use anyhow::Result;

pub fn read_line() -> Result<String> {
    let mut line = String::new();
    std::io::stdin().read_line(&mut line)?;
    Ok(line)
}

pub fn join_array<I, S>(arr: I) -> String
where
    I: IntoIterator<Item = S>,
    S: AsRef<str>,
{
    let mut iter = arr.into_iter();
    let Some(first) = iter.next() else {
        return "[]".to_string();
    };

    let mut joined = String::from("[");
    joined.push_str(first.as_ref());

    for item in iter {
        joined.push(',');
        joined.push_str(item.as_ref());
    }

    joined.push(']');
    joined
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_join_array_empty() {
        let arr: Vec<String> = vec![];
        assert_eq!(join_array(arr), "[]");
    }

    #[test]
    fn test_join_array_single_item() {
        assert_eq!(join_array(["1"]), "[1]");
    }

    #[test]
    fn test_join_array_multiple_items() {
        assert_eq!(join_array(["1", "2", "3"]), "[1,2,3]");
    }

    #[test]
    fn test_join_array_preserves_serialized_values() {
        let arr = [
            r#""hello""#.to_string(),
            "[1,2]".to_string(),
            r#"{"ok":true}"#.to_string(),
            "null".to_string(),
        ];

        assert_eq!(join_array(arr), r#"["hello",[1,2],{"ok":true},null]"#);
    }
}
