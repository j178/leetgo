#include <sstream>

#include "../LC_IO.h"

using namespace std;

int main_ret = 0;

template <typename T>
void test_scan_print(const char *raw) {
    std::stringstream in, out;
    in << raw;
    T x;
    LeetCodeIO::scan(in, x);
    LeetCodeIO::print(out, x);
    if (in.str() == out.str()) {
        printf("passed\n");
    } else {
        printf("want: %s, got: %s", raw, out.str().c_str());
        main_ret = 1;
    }
}

void test_all() {
    test_scan_print<int>("19890604");
    test_scan_print<int64_t>("1989060419890604");
    test_scan_print<bool>("true");
    test_scan_print<char>("\"a\"");
    test_scan_print<string>("\"he\\\"llo\"");
    test_scan_print<double>("1.98964");
    test_scan_print<ListNode*>("[19,89,0,6,0,4]");
    test_scan_print<TreeNode*>("[1989,null,6,null,4]");

    test_scan_print<vector<vector<int>>>("[[1989,6,4],[19890604],[]]");
    test_scan_print<vector<vector<int64_t>>>("[[1989060419890604,1989,6,4],[1989060419890604],[]]");
    //test_scan_print<vector<vector<bool>>("[[true,false,true],[false],[]]"); // https://isocpp.org/blog/2012/11/on-vectorbool
    test_scan_print<vector<vector<char>>>("[[\"t\",\"i\",\"a\",\"n\",\"a\",\"n\",\"m\",\"e\",\"n\"],[\"s\",\"q\",\"u\",\"a\",\"r\",\"e\"],[]]");
    test_scan_print<vector<vector<string>>>("[[\"tiananmen\",\"square\"],[""],[]]");
    test_scan_print<vector<vector<double>>>("[[1989.06040,19.89640],[1.98964],[]]");
    test_scan_print<vector<vector<ListNode*>>>("[[[19,89,0,6,0,4],[1989,6,4]],[[19890604]],[]]");
    test_scan_print<vector<vector<TreeNode*>>>("[[[1989,null,6,null,4],[1989,6,4]],[[19890604]],[]]");
}

int main() {
    test_all();
    return main_ret;
}
