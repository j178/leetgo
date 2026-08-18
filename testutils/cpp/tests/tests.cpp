#include "../LC_IO.h"

#include <cstdint>
#include <exception>
#include <iostream>
#include <sstream>
#include <string>
#include <vector>

namespace {

int failures = 0;

void fail(const std::string &message) {
    std::cerr << "FAILED: " << message << '\n';
    ++failures;
}

template <typename T>
void delete_nodes(const T &) {}

void delete_nodes(ListNode *const &head) {
    ListNode *node = head;
    while (node != nullptr) {
        ListNode *next = node->next;
        delete node;
        node = next;
    }
}

void delete_nodes(TreeNode *const &node) {
    if (node == nullptr) {
        return;
    }
    delete_nodes(node->left);
    delete_nodes(node->right);
    delete node;
}

template <typename T>
void delete_nodes(const std::vector<T> &values) {
    for (const auto &value : values) {
        delete_nodes(value);
    }
}

template <typename T>
void expect_serialized(const std::string &input_text, const std::string &expected) {
    std::ostringstream output;
    T value{};

    try {
        value = LeetCodeIO::deserialize<T>(input_text);
        LeetCodeIO::print(output, value);
    } catch (const std::exception &error) {
        fail(input_text + ": " + error.what());
        delete_nodes(value);
        return;
    }

    if (output.str() != expected) {
        fail(input_text + ": expected " + expected + ", got " + output.str());
    }
    delete_nodes(value);
}

template <typename T>
void expect_error(const std::string &input_text) {
    T value{};

    try {
        value = LeetCodeIO::deserialize<T>(input_text);
        delete_nodes(value);
        fail(input_text + ": expected an error");
    } catch (const LeetCodeIO::Error &) {
        delete_nodes(value);
    }
}

void test_scalars() {
    expect_serialized<int>("-19890604", "-19890604");
    expect_serialized<std::int64_t>("1989060419890604", "1989060419890604");
    expect_serialized<bool>("true", "true");
    expect_serialized<bool>("false", "false");
    expect_serialized<char>(R"("a")", R"("a")");
    expect_serialized<char>(R"("\\")", R"("\\")");
    expect_serialized<std::string>(R"("quote: \"; slash: \\ ")", R"("quote: \"; slash: \\ ")");
    expect_serialized<std::string>(R"("你好")", R"("你好")");
    expect_serialized<double>("1.5", "1.5");
}

void test_collections() {
    expect_serialized<std::vector<int>>(" [ 1, -2, 3 ] ", "[1,-2,3]");
    expect_serialized<std::vector<bool>>("[true,false,true]", "[true,false,true]");
    expect_serialized<std::vector<std::vector<int>>>("[[1,2],[],[3]]", "[[1,2],[],[3]]");
    expect_serialized<std::vector<std::string>>(R"(["a,b","[x]",""])", R"(["a,b","[x]",""])");
}

void test_nodes() {
    expect_serialized<ListNode *>("[]", "[]");
    expect_serialized<ListNode *>("[19,89,0,6,0,4]", "[19,89,0,6,0,4]");
    expect_serialized<TreeNode *>("[]", "[]");
    expect_serialized<TreeNode *>("[1,2,3,null,4]", "[1,2,3,null,4]");
    expect_serialized<TreeNode *>("[1,null,null,null]", "[1]");
    expect_serialized<std::vector<ListNode *>>("[[1,2],[],[3]]", "[[1,2],[],[3]]");
    expect_serialized<std::vector<TreeNode *>>("[[1,null,2],[],[3]]", "[[1,null,2],[],[3]]");
}

void test_invalid_input() {
    expect_error<std::vector<int>>("[1,2,3");
    expect_error<std::vector<int>>("[1,null,3]");
    expect_error<std::vector<int>>("[1,2,]");
    expect_error<std::vector<int>>("1,2,3");
    expect_error<bool>("True");
    expect_error<char>(R"("ab")");
    expect_error<std::string>(R"("unterminated)");
}

void test_split_array() {
    const std::vector<std::string> values = LeetCodeIO::split_array(
        R"( [1,["a,b",2],{"items":[true,false]},"a\",b",null] )"
    );
    const std::vector<std::string> expected = {
        "1",
        R"(["a,b",2])",
        R"({"items":[true,false]})",
        R"("a\",b")",
        "null",
    };
    if (values != expected) {
        fail("split array returned unexpected elements");
    }

    try {
        LeetCodeIO::split_array("[1,]");
        fail("split array accepted a trailing comma");
    } catch (const LeetCodeIO::Error &) {
    }
}

void test_serialized_output() {
    std::vector<std::string> values = {
        "null",
        LeetCodeIO::serialize(std::string("text")),
        LeetCodeIO::serialize(std::vector<int>{1, 2}),
        LeetCodeIO::serialize(true),
    };
    const std::string output = LeetCodeIO::join_array(values);
    if (output != R"([null,"text",[1,2],true])") {
        fail("serialized output: got " + output);
    }
}

void test_legacy_vector_overloads() {
    std::istringstream input("[1,2]");
    std::vector<int> values;
    LeetCodeIO::scan<int>(input, values);

    std::ostringstream output;
    LeetCodeIO::print<int>(output, values);
    if (output.str() != "[1,2]") {
        fail("legacy vector overloads: got " + output.str());
    }
}

}  // namespace

int main() {
    test_scalars();
    test_collections();
    test_nodes();
    test_invalid_input();
    test_split_array();
    test_serialized_output();
    test_legacy_vector_overloads();

    if (failures == 0) {
        std::cout << "all C++ testutils tests passed\n";
    }
    return failures == 0 ? 0 : 1;
}
