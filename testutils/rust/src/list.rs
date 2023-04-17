use serde::{Deserialize, Deserializer, Serialize, Serializer};
use serde::de::Visitor;
use serde::ser::SerializeSeq;

type ListLink = Option<Box<ListNode>>;

#[derive(Debug, PartialEq, Eq, Clone)]
pub struct ListNode {
    pub val: i32,
    pub next: ListLink,
}

#[macro_export]
macro_rules! list {
    () => {
        None
    };
    ($e:expr) => {
        Some(Box::new(ListNode {
            val: $e,
            next: None,
        }))
    };
    ($e:expr, $($tail:tt)*) => {
        Some(Box::new(ListNode {
            val: $e,
            next: list!($($tail)*),
        }))
    };
}

#[derive(Debug, PartialEq, Eq, Clone)]
pub struct LinkedList(ListLink);

impl From<LinkedList> for Option<Box<ListNode>> {
    fn from(list: LinkedList) -> Self {
        list.0
    }
}

impl Serialize for LinkedList {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
        where
            S: Serializer,
    {
        let mut seq = serializer.serialize_seq(None)?;
        let mut current = &self.0;
        while let Some(ref node) = current {
            seq.serialize_element(&node.val)?;
            current = &node.next;
        }
        seq.end()
    }
}

struct LinkedListVisitor;

impl<'de> Visitor<'de> for LinkedListVisitor {
    type Value = LinkedList;

    fn expecting(&self, formatter: &mut std::fmt::Formatter) -> std::fmt::Result {
        formatter.write_str("a list of integers")
    }

    fn visit_seq<A>(self, mut seq: A) -> Result<Self::Value, A::Error>
        where
            A: serde::de::SeqAccess<'de>,
    {
        let mut head = None;
        let mut current = &mut head;
        while let Some(val) = seq.next_element()? {
            let node = ListNode { val, next: None };
            *current = Some(Box::new(node));
            current = &mut current.as_mut().unwrap().next;
        }
        Ok(LinkedList(head))
    }
}

impl<'de> Deserialize<'de> for LinkedList {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
        where
            D: Deserializer<'de>,
    {
        deserializer.deserialize_seq(LinkedListVisitor)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_list_serialize() {
        let list = LinkedList(list!(1, 2, 3));
        let serialized = serde_json::to_string(&list).unwrap();
        assert_eq!(serialized, "[1,2,3]");
    }

    #[test]
    fn test_list_deserialize() {
        let serialized = "[1,2,3]";
        let list: LinkedList = serde_json::from_str(serialized).unwrap();
        assert_eq!(list, LinkedList(list![1, 2, 3]));

        let serialized = "[]";
        let list: LinkedList = serde_json::from_str(serialized).unwrap();
        assert!(list.0.is_none());

        let serialized = "[true]";
        let list = serde_json::from_str::<LinkedList>(serialized);
        assert!(list.is_err());
    }
}
