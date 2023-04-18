#ifndef LC_IO_H
#define LC_IO_H

#include <iomanip>
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
 * Definition for a binary tree node.
 */
struct TreeNode {
    int val;
    TreeNode *left;
    TreeNode *right;
    TreeNode(int x) : val(x), left(nullptr), right(nullptr) {}
    TreeNode(int x, TreeNode *left, TreeNode *right) : val(x), left(left), right(right) {}
};

namespace LeetCodeIO {
    namespace Helper {
        /**
         * Function for deserializing a singly-linked list.
         */
        inline void scan_list(std::istream &is, ListNode *&node) {
            node = nullptr;
            ListNode *now = nullptr;
        [[maybe_unused]]
        L0: is.ignore();
        L1: switch (is.peek()) {
            case ' ':
            case ',': is.ignore(); goto L1;
            case ']': is.ignore(); goto L2;
            default : int x; is >> x;
                      now = (now ? now->next : node) = new ListNode(x);
                      goto L1;
            }
        L2: return;
        }

        /**
         * Function for deserializing a binary tree.
         */
        inline void scan_tree(std::istream &is, TreeNode *&node) {
            std::deque<TreeNode *> dq;
        [[maybe_unused]]
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
        L2: int n = dq.size();
            for (int i = 0, j = 1; i < n; ++i) {
                auto root = dq[i];
                if (root == nullptr) { continue; }
                root->left = j < n ? dq[j] : nullptr;
                root->right = j + 1 < n ? dq[j + 1] : nullptr;
                j += 2;
            }
            node = n ? dq[0] : nullptr;
            return;
        }

        /**
         * Function for serializing a singly-linked list.
         */
        inline void print_list(std::ostream &os, ListNode *node) {
            os.put('[');
            if (node != nullptr) {
                do {
                    os << node->val; os.put(',');
                    node = node->next;
                } while(node != nullptr);
                os.seekp(-1, std::ios_base::end);
            }
            os.put(']');
            return;
        }

        /**
         * Function for serializing a binary tree.
         */
        inline void print_tree(std::ostream &os, TreeNode *node) {
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
                    os << front->val; os.put(',');
                } else {
                    os << "null,";
                }
            };
            os.put('[');
            if (node != nullptr) {
                push(node);
                while (cnt_not_null_nodes > 0) { pop(); }
                os.seekp(-1, std::ios_base::end);
            }
            os.put(']');
            return;
        }
    }

    /**
     * Function for scanning a variable.
     */
    template<typename T>
    void scan(std::istream &is, T &x) {
        /**
         * operator >> discards leading whitespaces by default
         * when not using operator >>, they must be discarded explicitly
         */
        if constexpr (std::is_same_v<T, std::string>) {
            is >> std::quoted(x);
        } else if constexpr (std::is_same_v<T, bool>) {
            is >> std::ws; x = is.get() == 't'; is.ignore(4 - x);
        } else if constexpr (std::is_same_v<T, char>) {
            is >> std::ws; is.ignore(); x = is.get(); is.ignore();
        } else if constexpr (std::is_same_v<T, ListNode *>) {
            is >> std::ws; Helper::scan_list(is, x);
        } else if constexpr (std::is_same_v<T, TreeNode *>) {
            is >> std::ws; Helper::scan_tree(is, x);
        } else {
            is >> x;
        }
    }

    /**
     * Function for deserializing an array.
     */
    template <typename T>
    void scan(std::istream &is, std::vector<T> &v) {
    [[maybe_unused]]
    L0: is >> std::ws;
        is.ignore();
    L1: switch (is.peek()) {
        case ' ':
        case ',': is.ignore(); goto L1;
        case ']': is.ignore(); goto L2;
        default : v.emplace_back();
                  scan(is, v.back());
                  goto L1;
        }
    L2: return;
    }

    /**
     * Function for printing a variable.
     */
    template<typename T>
    void print(std::ostream &os, const T& x) {
        if constexpr (std::is_same_v<T, std::string>) {
            os.put('"'); os << x; os.put('"'); 
        } else if constexpr (std::is_same_v<T, double>) {
            constexpr int siz = 320;
            char buf[siz]; snprintf(buf, siz, "%.5f", x); os << buf;
        } else if constexpr (std::is_same_v<T, bool>) {
            static const char tab[2][8] = {"false", "true"};
            os.write(tab[x], x ? 4 : 5);
        } else if constexpr (std::is_same_v<T, char>) {
            os.put('"'); os.put(x); os.put('"');
        } else if constexpr (std::is_same_v<T, ListNode *>) {
            Helper::print_list(os, x);
        } else if constexpr (std::is_same_v<T, TreeNode *>) {
            Helper::print_tree(os, x);
        } else {
            os << x;
        }
    }

    /**
     * Function for serializing an array.
     */
    template <typename T>
    void print(std::ostream &os, const std::vector<T> &v) {
        os.put('[');
        for (auto &&x : v) {
            print(os, x);
            os.put(',');
        }
        os.seekp(-!v.empty(), std::ios_base::end);
        os.put(']');
    }
};

#endif
