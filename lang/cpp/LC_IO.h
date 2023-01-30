auto &operator>>(istream &is, ListNode *&node) {
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

auto &operator<<(ostream &os, ListNode *node) {
	os << '[';
	while (node != nullptr) {
		os << node->val << ',';
		node = node->next;
	}
	os.seekp(-1, ios_base::end);
	os << ']';
	return os;
}

auto &operator>>(istream &is, TreeNode *&node) {
	deque<TreeNode *> dq;
L0: is.ignore();
L1: switch (is.peek()) {
	case ' ':
	case ',': is.ignore(); goto L1;
	case 'n': is.ignore(5); dq.emplace_back(nullptr);
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

auto &operator<<(ostream &os, TreeNode *node) {
	queue<TreeNode *> q;
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
		while (cnt_not_null_nodes > 0) { pop();	}
		os.seekp(-1, ios_base::end);
	}
	os << ']';
	return os;
}

template <typename T>
auto &operator>>(istream &is, vector<T> &v) {
L0: is.ignore();
L1: switch (is.peek()) {
	case ' ':
	case ',': is.ignore(); goto L1;
	case ']': is.ignore(); goto L2;
	default : v.emplace_back();
	          if constexpr (is_same_v<T, string>) {
	              is >> quoted(v.back());
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

template <typename T>
auto &operator<<(ostream &os, const vector<T> &v){
	os << '[';
	if constexpr (is_same_v<T, string>) {
		for (auto &&x : v) { os << quoted(x) << ','; }
	} else if constexpr (is_same_v<T, double>) {
		for (auto &&x : v) {
			char buf[320]; sprintf(buf, "%.5f,", x); os << buf;
		}
	} else {
		for (auto &&x : v) { os << x << ','; }
	}
	os.seekp(-1, ios_base::end);
	os << ']';
	return os;
}
