#include <sstream>

#include <cassert>

#include "../LC_IO.h"

#define LC_IO_TEST_MACRO(TYP, STR) \
test<TYP>(STR, #TYP)

template <typename T>
void test(const char *raw, const char *t) {
    std::stringstream in, out;
    in << raw;
    T x;
    LeetCodeIO::scan(in, x);
    LeetCodeIO::print(out, x);
    printf("[LC_IO_TEST] ser / deser <%s> in:'%s' out:'%s' ", t, raw, out.str().c_str());
    assert(in.str() == out.str());
    printf("passed\n");
}

inline void test_all() {
    LC_IO_TEST_MACRO(int, "19890604");
    LC_IO_TEST_MACRO(int64_t, "1989060419890604");
    LC_IO_TEST_MACRO(bool, "true");
    LC_IO_TEST_MACRO(char, "\"a\"");
    LC_IO_TEST_MACRO(std::string, "\"hello\"");
    LC_IO_TEST_MACRO(double, "1.98964");
    LC_IO_TEST_MACRO(ListNode *, "[19,89,0,6,0,4]");
    LC_IO_TEST_MACRO(TreeNode *, "[1989,null,6,null,4]");

    LC_IO_TEST_MACRO(std::vector<std::vector<int>>, "[[1989,6,4],[19890604],[]]");
    LC_IO_TEST_MACRO(std::vector<std::vector<int64_t>>, "[[1989060419890604,1989,6,4],[1989060419890604],[]]");
    //LC_IO_TEST_MACRO(std::vector<std::vector<bool>>, "[[true,false,true],[false],[]]"); // https://isocpp.org/blog/2012/11/on-vectorbool
    LC_IO_TEST_MACRO(std::vector<std::vector<char>>, "[[\"t\",\"i\",\"a\",\"n\",\"a\",\"n\",\"m\",\"e\",\"n\"],[\"s\",\"q\",\"u\",\"a\",\"r\",\"e\"],[]]");
    LC_IO_TEST_MACRO(std::vector<std::vector<std::string>>, "[[\"tiananmen\",\"square\"],[""],[]]");
    LC_IO_TEST_MACRO(std::vector<std::vector<double>>, "[[1989.06040,19.89640],[1.98964],[]]");
    LC_IO_TEST_MACRO(std::vector<std::vector<ListNode *>>, "[[[19,89,0,6,0,4],[1989,6,4]],[[19890604]],[]]");
    LC_IO_TEST_MACRO(std::vector<std::vector<TreeNode *>>, "[[[1989,null,6,null,4],[1989,6,4]],[[19890604]],[]]");
}

int main() {
    test_all();
    return 0;
}
