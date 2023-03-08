#ifndef LC_IO_H
#define LC_IO_H

#include <iostream>
#include <queue>

/**
 * Definition for a singly-linked list.
 */
struct ListNode {
    int val;
    ListNode *next;
    ListNode() : val(0), next(nullptr) {}
    ListNode(int x) : val(x), next(nullptr) {}
    ListNode(int x, ListNode *next) : val(x), next(next) {}
};

/**
 * Function for deserializing a singly-linked list.
 */
std::istream &operator>>(std::istream &is, ListNode *&node) {
    node = nullptr;
    ListNode *now = nullptr;
L0: is.ignore();
L1: switch (is.peek()) {
    case ' ':
    case ',': is.ignore(); goto L1;
    case ']': is.ignore(); goto L2;
    default : int x; is >> x;
              now = (now ? now->next : node) = new ListNode(x);
              goto L1;
    }
L2: switch (is.peek()) {
    case '\r': is.ignore(); goto L2;
    case '\n': is.ignore(); goto L3;
    case EOF : goto L3;
    }
L3: return is;
}

/**
 * Function for serializing a singly-linked list.
 */
std::ostream &operator<<(std::ostream &os, ListNode *node) {
    os << '[';
    if (node != nullptr) {
        do {
            os << node->val << ',';
            node = node->next;
        } while(node != nullptr);
        os.seekp(-1, std::ios_base::end);
    }
    os << ']';
    return os;
}

/**
 * Definition for a binary tree node.
 */
struct TreeNode {
    int val;
    TreeNode *left;
    TreeNode *right;
    TreeNode(int x) : val(x), left(nullptr), right(nullptr) {}
    TreeNode(int x, TreeNode *left, TreeNode *right) : val(x), left(left), right(right) {}
};

/**
 * Function for deserializing a binary tree.
 */
std::istream &operator>>(std::istream &is, TreeNode *&node) {
    std::deque<TreeNode *> dq;
L0: is.ignore();
L1: switch (is.peek()) {
    case ' ':
    case ',': is.ignore(); goto L1;
    case 'n': is.ignore(4); dq.emplace_back(nullptr);
              goto L1;
    case ']': is.ignore(); goto L2;
    default : int x; is >> x;
              dq.emplace_back(new TreeNode(x));
              goto L1;
    }
L2: switch (is.peek()) {
    case '\r': is.ignore(); goto L2;
    case '\n': is.ignore(); goto L3;
    case EOF : goto L3;
    }
L3: int n = dq.size();
    for (int i = 0, j = 1; i < n; ++i) {
        auto root = dq[i];
        if (root == nullptr) { continue; }
        root->left = j < n ? dq[j] : nullptr;
        root->right = j + 1 < n ? dq[j + 1] : nullptr;
        j += 2;
    }
    node = n ? dq[0] : nullptr;
    return is;
}

/**
 * Function for serializing a binary tree.
 */
std::ostream &operator<<(std::ostream &os, TreeNode *node) {
    std::queue<TreeNode *> q;
    int cnt_not_null_nodes = 0;
    auto push = [&](TreeNode *node) {
        q.emplace(node);
        if (node != nullptr) { ++cnt_not_null_nodes; }
    };
    auto pop = [&]() {
        auto front = q.front(); q.pop();
        if (front != nullptr) {
            --cnt_not_null_nodes;
            push(front->left);
            push(front->right);
            os << front->val << ',';
        } else {
            os << "null,";
        }
    };
    os << '[';
    if (node != nullptr) {
        push(node);
        while (cnt_not_null_nodes > 0) { pop(); }
        os.seekp(-1, std::ios_base::end);
    }
    os << ']';
    return os;
}

/**
 * Function for deserializing an array.
 */
template <typename T>
std::istream &operator>>(std::istream &is, std::vector<T> &v) {
L0: is.ignore();
L1: switch (is.peek()) {
    case ' ':
    case ',': is.ignore(); goto L1;
    case ']': is.ignore(); goto L2;
    default : v.emplace_back();
              if constexpr (std::is_same_v<T, std::string>) {
                  is >> quoted(v.back());
              } else if constexpr (std::is_same_v<T, bool>) {
                  bool t = is.get() == 't'; v.back() = t; is.ignore(4 - t);
              } else if constexpr (std::is_same_v<T, char>) {
                  is.ignore(); v.back() = is.get(); is.ignore();
              } else {
                  is >> v.back();
              }
              goto L1;
    }
L2: switch (is.peek()) {
    case '\r': is.ignore(); goto L2;
    case '\n': is.ignore(); goto L3;
    case EOF : goto L3;
}
L3: return is;
}

/**
 * Function for serializing an array.
 */
template <typename T>
std::ostream &operator<<(std::ostream &os, const std::vector<T> &v) {
    os << '[';
    for (auto &&x : v) {
        if constexpr (std::is_same_v<T, std::string>) {
            os << quoted(x) << ',';
        } else if constexpr (std::is_same_v<T, double>) {
            char buf[320]; sprintf(buf, "%.5f,", x); os << buf;
        } else if constexpr (std::is_same_v<T, bool>) {
            const char *buf = "false,\0\0true,"; os << buf + (x << 3);
        } else if constexpr (std::is_same_v<T, char>) {
            os << '"' << x << "\",";
        } else {
            os << x << ',';
        }
    }
    os.seekp(-!v.empty(), std::ios_base::end);
    os << ']';
    return os;
}

#endif
