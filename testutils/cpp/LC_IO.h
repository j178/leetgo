#ifndef LEETGO_LC_IO_H
#define LEETGO_LC_IO_H

#include <cctype>
#include <cstddef>
#include <iomanip>
#include <istream>
#include <limits>
#include <optional>
#include <ostream>
#include <queue>
#include <sstream>
#include <stdexcept>
#include <string>
#include <type_traits>
#include <utility>
#include <vector>

struct ListNode {
    int val;
    ListNode *next;

    ListNode() : val(0), next(nullptr) {}
    ListNode(int value) : val(value), next(nullptr) {}
    ListNode(int value, ListNode *next_node) : val(value), next(next_node) {}
};

struct TreeNode {
    int val;
    TreeNode *left;
    TreeNode *right;

    TreeNode() : val(0), left(nullptr), right(nullptr) {}
    TreeNode(int value) : val(value), left(nullptr), right(nullptr) {}
    TreeNode(int value, TreeNode *left_node, TreeNode *right_node)
        : val(value), left(left_node), right(right_node) {}
};

namespace LeetCodeIO {

// New API: parsing failures reported by deserialize and split_array.
class Error : public std::runtime_error {
public:
    using std::runtime_error::runtime_error;
};

// Implementation details. Generated code must not use names from this namespace.
namespace detail {

inline void skip_whitespace(std::istream &input) {
    input >> std::ws;
}

inline int peek(std::istream &input) {
    skip_whitespace(input);
    return input.peek();
}

inline bool consume(std::istream &input, char expected) {
    if (peek(input) != expected) {
        return false;
    }
    input.get();
    return true;
}

inline void read_literal(std::istream &input, const char *literal) {
    skip_whitespace(input);
    for (const char *p = literal; *p != '\0'; ++p) {
        if (input.get() != *p) {
            throw Error(std::string("expected '") + literal + "'");
        }
    }
}

inline void expect(std::istream &input, char expected) {
    skip_whitespace(input);
    if (input.get() != expected) {
        throw Error(std::string("expected '") + expected + "'");
    }
}

inline void expect_end(std::istream &input) {
    skip_whitespace(input);
    if (input.peek() != std::char_traits<char>::eof()) {
        throw Error("unexpected trailing input");
    }
}

template <typename T>
void scan(std::istream &input, T &value);

template <typename T>
void scan(std::istream &input, std::vector<T> &values);

template <typename T>
void scan(std::istream &input, std::optional<T> &value);

inline void scan(std::istream &input, std::string &value);
inline void scan(std::istream &input, bool &value);
inline void scan(std::istream &input, char &value);
inline void scan(std::istream &input, ListNode *&node);
inline void scan(std::istream &input, TreeNode *&node);

template <typename T>
void scan(std::istream &input, T &value) {
    static_assert(std::is_arithmetic_v<T>, "unsupported input type");
    skip_whitespace(input);
    if (!(input >> value)) {
        throw Error("expected a number");
    }
}

inline void scan(std::istream &input, std::string &value) {
    if (peek(input) != '"' || !(input >> std::quoted(value))) {
        throw Error("expected a quoted string");
    }
}

inline void scan(std::istream &input, bool &value) {
    const int next = peek(input);
    if (next == 't') {
        read_literal(input, "true");
        value = true;
    } else if (next == 'f') {
        read_literal(input, "false");
        value = false;
    } else {
        throw Error("expected a boolean");
    }
}

inline void scan(std::istream &input, char &value) {
    std::string text;
    scan(input, text);
    if (text.size() != 1) {
        throw Error("expected a single-byte character");
    }
    value = text[0];
}

template <typename T>
void scan(std::istream &input, std::optional<T> &value) {
    if (peek(input) == 'n') {
        read_literal(input, "null");
        value.reset();
        return;
    }

    value.emplace();
    scan(input, *value);
}

template <typename T>
void scan(std::istream &input, std::vector<T> &values) {
    expect(input, '[');
    values.clear();

    while (!consume(input, ']')) {
        if (!values.empty()) {
            expect(input, ',');
        }
        T value{};
        scan(input, value);
        values.push_back(std::move(value));
    }
}

inline void scan(std::istream &input, ListNode *&node) {
    std::vector<int> values;
    scan(input, values);

    ListNode dummy;
    ListNode *tail = &dummy;
    for (int value : values) {
        tail->next = new ListNode(value);
        tail = tail->next;
    }
    node = dummy.next;
}

inline void scan(std::istream &input, TreeNode *&node) {
    std::vector<std::optional<int>> values;
    scan(input, values);

    if (values.empty()) {
        node = nullptr;
        return;
    }

    if (!values[0].has_value()) {
        node = nullptr;
        return;
    }

    TreeNode *root = new TreeNode(*values[0]);
    std::queue<TreeNode *> parents;
    parents.push(root);
    std::size_t index = 1;
    while (index < values.size() && !parents.empty()) {
        TreeNode *parent = parents.front();
        parents.pop();

        if (values[index].has_value()) {
            parent->left = new TreeNode(*values[index]);
            parents.push(parent->left);
        }
        ++index;

        if (index < values.size() && values[index].has_value()) {
            parent->right = new TreeNode(*values[index]);
            parents.push(parent->right);
        }
        ++index;
    }
    node = root;
}

inline std::string read_line(std::istream &input) {
    std::string line;
    if (!std::getline(input >> std::ws, line)) {
        throw Error("expected an input line");
    }
    return line;
}

inline std::vector<std::string> split_array(const std::string &text) {
    const auto trim = [](const std::string &value) {
        std::size_t begin = 0;
        while (begin < value.size() &&
               std::isspace(static_cast<unsigned char>(value[begin]))) {
            ++begin;
        }

        std::size_t end = value.size();
        while (end > begin &&
               std::isspace(static_cast<unsigned char>(value[end - 1]))) {
            --end;
        }
        return value.substr(begin, end - begin);
    };

    const std::string array = trim(text);
    if (array.size() < 2 || array.front() != '[' || array.back() != ']') {
        throw Error("expected an array");
    }

    const std::string body = array.substr(1, array.size() - 2);
    std::vector<std::string> values;
    if (trim(body).empty()) {
        return values;
    }

    std::size_t start = 0;
    std::string closing_delimiters;
    bool in_string = false;
    bool escaped = false;
    for (std::size_t i = 0; i < body.size(); ++i) {
        const char ch = body[i];
        if (in_string) {
            if (escaped) {
                escaped = false;
            } else if (ch == '\\') {
                escaped = true;
            } else if (ch == '"') {
                in_string = false;
            }
            continue;
        }

        if (ch == '"') {
            in_string = true;
        } else if (ch == '[' || ch == '{') {
            closing_delimiters.push_back(ch == '[' ? ']' : '}');
        } else if (ch == ']' || ch == '}') {
            if (closing_delimiters.empty() || closing_delimiters.back() != ch) {
                throw Error("invalid array element");
            }
            closing_delimiters.pop_back();
        } else if (ch == ',' && closing_delimiters.empty()) {
            const std::string value = trim(body.substr(start, i - start));
            if (value.empty()) {
                throw Error("expected an array element");
            }
            values.push_back(value);
            start = i + 1;
        }
    }

    if (in_string || !closing_delimiters.empty()) {
        throw Error("unterminated array element");
    }

    const std::string value = trim(body.substr(start));
    if (value.empty()) {
        throw Error("expected an array element");
    }
    values.push_back(value);
    return values;
}

template <typename T>
T deserialize(const std::string &text) {
    std::istringstream input(text);
    T value{};
    scan(input, value);
    expect_end(input);
    return value;
}

template <typename T>
T deserialize(const std::vector<std::string> &values, std::size_t index) {
    if (index >= values.size()) {
        throw Error("missing array element");
    }
    return deserialize<T>(values[index]);
}

template <typename T>
void print(std::ostream &output, const T &value);

template <typename T>
void print(std::ostream &output, const std::vector<T> &values);

inline void print(std::ostream &output, const std::string &value);
inline void print(std::ostream &output, bool value);
inline void print(std::ostream &output, char value);
inline void print(std::ostream &output, double value);
inline void print(std::ostream &output, const ListNode *node);
inline void print(std::ostream &output, ListNode *node);
inline void print(std::ostream &output, const TreeNode *node);
inline void print(std::ostream &output, TreeNode *node);

template <typename T>
void print(std::ostream &output, const T &value) {
    static_assert(std::is_arithmetic_v<T>, "unsupported output type");
    output << value;
}

inline void print(std::ostream &output, const std::string &value) {
    output << std::quoted(value);
}

inline void print(std::ostream &output, bool value) {
    output << (value ? "true" : "false");
}

inline void print(std::ostream &output, char value) {
    print(output, std::string(1, value));
}

inline void print(std::ostream &output, double value) {
    const std::streamsize old_precision = output.precision();
    output << std::setprecision(std::numeric_limits<double>::max_digits10) << value;
    output.precision(old_precision);
}

template <typename T>
void print(std::ostream &output, const std::vector<T> &values) {
    output.put('[');
    for (std::size_t i = 0; i < values.size(); ++i) {
        if (i > 0) {
            output.put(',');
        }
        if constexpr (std::is_same_v<T, bool>) {
            print(output, static_cast<bool>(values[i]));
        } else {
            print(output, values[i]);
        }
    }
    output.put(']');
}

inline void print(std::ostream &output, const ListNode *node) {
    output.put('[');
    bool first = true;
    while (node != nullptr) {
        if (!first) {
            output.put(',');
        }
        print(output, node->val);
        first = false;
        node = node->next;
    }
    output.put(']');
}

inline void print(std::ostream &output, ListNode *node) {
    print(output, static_cast<const ListNode *>(node));
}

inline void print(std::ostream &output, const TreeNode *node) {
    if (node == nullptr) {
        output << "[]";
        return;
    }

    std::vector<const TreeNode *> values;
    std::queue<const TreeNode *> pending;
    pending.push(node);

    while (!pending.empty()) {
        const TreeNode *current = pending.front();
        pending.pop();
        values.push_back(current);

        if (current == nullptr) {
            continue;
        }
        pending.push(current->left);
        pending.push(current->right);
    }

    while (!values.empty() && values.back() == nullptr) {
        values.pop_back();
    }

    output.put('[');
    for (std::size_t i = 0; i < values.size(); ++i) {
        if (i > 0) {
            output.put(',');
        }
        if (values[i] == nullptr) {
            output << "null";
        } else {
            print(output, values[i]->val);
        }
    }
    output.put(']');
}

inline void print(std::ostream &output, TreeNode *node) {
    print(output, static_cast<const TreeNode *>(node));
}

}  // namespace detail

// Legacy compatibility API.
//
// Older generated code calls scan and print directly. Keep both the generic and
// vector overloads source-compatible. Current non-system-design code also uses
// print, but newly generated code uses deserialize instead of scan.
template <typename T>
void scan(std::istream &input, T &value) {
    detail::scan(input, value);
}

template <typename T>
void scan(std::istream &input, std::vector<T> &values) {
    detail::scan(input, values);
}

template <typename T>
void print(std::ostream &output, const T &value) {
    detail::print(output, value);
}

template <typename T>
void print(std::ostream &output, const std::vector<T> &values) {
    detail::print(output, values);
}

// New API used by the current generator.
//
// deserialize reads either one complete input line or one previously split
// system-design argument. split_array separates the top-level elements of a
// system-design parameter list. serialize and join_array build output values.
template <typename T>
T deserialize(const std::string &text) {
    return detail::deserialize<T>(text);
}

template <typename T>
T deserialize(std::istream &input) {
    return detail::deserialize<T>(detail::read_line(input));
}

template <typename T>
T deserialize(const std::vector<std::string> &values, std::size_t index) {
    return detail::deserialize<T>(values, index);
}

inline std::vector<std::string> split_array(const std::string &text) {
    return detail::split_array(text);
}

inline std::vector<std::string> split_array(std::istream &input) {
    return detail::split_array(detail::read_line(input));
}

template <typename T>
std::string serialize(const T &value) {
    std::ostringstream output;
    detail::print(output, value);
    return output.str();
}

inline std::string join_array(const std::vector<std::string> &values) {
    std::ostringstream output;
    output.put('[');
    for (std::size_t i = 0; i < values.size(); ++i) {
        if (i > 0) {
            output.put(',');
        }
        output << values[i];
    }
    output.put(']');
    return output.str();
}

}  // namespace LeetCodeIO

#endif
