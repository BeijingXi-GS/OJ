// Package problems is the trusted, compile-time catalog for the local MVP.
// Files on disk do not register problems, and user input never supplies paths.
package problems

type Sample struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type Definition struct {
	ID                string            `json:"id"`
	Title             string            `json:"title"`
	Summary           string            `json:"summary"`
	Difficulty        string            `json:"difficulty"`
	Description       string            `json:"description"`
	InputDescription  string            `json:"input_description"`
	OutputDescription string            `json:"output_description"`
	Notes             string            `json:"notes"`
	EngineeringNotes  string            `json:"engineering_notes,omitempty"`
	Samples           []Sample          `json:"samples"`
	StarterCode       string            `json:"starter_code"`
	StarterCodes      map[string]string `json:"starter_codes,omitempty"`
	Type              string            `json:"type"`
}

var catalog = []Definition{
	{
		ID: "a-plus-b", Title: "A + B", Summary: "读取两个整数并输出它们的和。", Difficulty: "入门",
		Description:       "读取两个整数 a 和 b，计算并输出它们的和。",
		InputDescription:  "输入一行，包含两个以空格分隔的整数 a、b，范围均为 −10⁹ ≤ a, b ≤ 10⁹。",
		OutputDescription: "输出一个整数，表示 a + b。",
		Notes:             "请使用 long long 保存整数。",
		EngineeringNotes:  "默认代码是可运行的演示解法。用例按空白分隔的 token 比较，末尾空格或换行不影响正确性。",
		Samples:           []Sample{{Input: "1 2\n", Output: "3\n"}},
		StarterCode:       "#include <iostream>\n\nint main() {\n    long long a, b;\n    std::cin >> a >> b;\n    std::cout << a + b << '\\n';\n    return 0;\n}\n",
	},
	{
		ID: "array-sum", Title: "数组元素和", Summary: "计算一组整数的总和，注意结果可能超过 32 位整数范围。", Difficulty: "入门",
		Description:       "给定 n 个整数，计算它们的总和。数组中可能有负数和零。",
		InputDescription:  "第一行包含整数 n（1 ≤ n ≤ 1000）。第二行包含 n 个整数 ai（−10⁹ ≤ ai ≤ 10⁹），以空白分隔。",
		OutputDescription: "输出一个整数，表示所有数组元素的和。",
		Notes:             "总和的绝对值最大为 10¹²，必须使用 long long。",
		EngineeringNotes:  "默认代码是可运行的演示解法。用例按空白分隔的 token 比较，末尾空格或换行不影响正确性。",
		Samples:           []Sample{{Input: "5\n1 2 3 4 5\n", Output: "15\n"}, {Input: "4\n-5 0 2 -1\n", Output: "-4\n"}},
		StarterCode:       "#include <iostream>\n\nint main() {\n    int n;\n    std::cin >> n;\n    long long sum = 0;\n    for (int i = 0; i < n; ++i) {\n        long long value;\n        std::cin >> value;\n        sum += value;\n    }\n    std::cout << sum << '\\n';\n    return 0;\n}\n",
	},
	{
		ID: "maximum-element", Title: "数组最大值", Summary: "在包含负数、重复元素的一组整数中找出最大值。", Difficulty: "入门",
		Description:       "给定 n 个整数，输出其中最大的整数。最大值可能是负数，数组中也可能有重复元素。",
		InputDescription:  "第一行包含整数 n（1 ≤ n ≤ 1000）。第二行包含 n 个整数 ai（−10⁹ ≤ ai ≤ 10⁹），以空白分隔。",
		OutputDescription: "输出一个整数，表示数组中的最大值。",
		Notes:             "用第一个元素初始化最大值，不能默认从 0 开始。",
		EngineeringNotes:  "默认代码是可运行的演示解法。用例按空白分隔的 token 比较，末尾空格或换行不影响正确性。",
		Samples:           []Sample{{Input: "6\n3 1 9 -2 9 4\n", Output: "9\n"}, {Input: "3\n-8 -3 -5\n", Output: "-3\n"}},
		StarterCode:       "#include <iostream>\n\nint main() {\n    int n;\n    std::cin >> n;\n    long long best;\n    std::cin >> best;\n    for (int i = 1; i < n; ++i) {\n        long long value;\n        std::cin >> value;\n        if (value > best) best = value;\n    }\n    std::cout << best << '\\n';\n    return 0;\n}\n",
	},
	{
		ID: "palindrome", Title: "判断回文", Summary: "判断一个小写字母字符串从前后读取是否相同。", Difficulty: "入门",
		Description:       "给定一个只包含小写英文字母的非空字符串。如果从左向右和从右向左读取的内容相同，它就是回文。",
		InputDescription:  "输入一行字符串 s，仅包含 a 到 z，长度为 1 到 1000。",
		OutputDescription: "如果 s 是回文，输出大写 YES；否则输出大写 NO。",
		Notes:             "单个字符也属于回文；奇数和偶数长度都需要正确处理。",
		EngineeringNotes:  "默认代码是可运行的演示解法。用例按空白分隔的 token 比较，末尾空格或换行不影响正确性。",
		Samples:           []Sample{{Input: "abba\n", Output: "YES\n"}, {Input: "abca\n", Output: "NO\n"}},
		StarterCode:       "#include <algorithm>\n#include <iostream>\n#include <string>\n\nint main() {\n    std::string s;\n    std::cin >> s;\n    std::string reversed = s;\n    std::reverse(reversed.begin(), reversed.end());\n    std::cout << (s == reversed ? \"YES\" : \"NO\") << '\\n';\n    return 0;\n}\n",
	},
	{
		ID: "sort-numbers", Title: "整数排序", Summary: "将一组含负数或重复值的整数按升序排列。", Difficulty: "入门",
		Description:       "给定 n 个整数，将它们按从小到大的顺序输出。重复的整数必须保留。",
		InputDescription:  "第一行包含整数 n（1 ≤ n ≤ 1000）。第二行包含 n 个整数 ai（−10⁹ ≤ ai ≤ 10⁹），以空白分隔。",
		OutputDescription: "输出 n 个整数，按非递减顺序排列，以空白分隔。",
		Notes:             "注意保留重复元素。",
		EngineeringNotes:  "用例按空白分隔的 token 比较，末尾空格或换行不影响正确性。默认代码是可运行的演示解法。",
		Samples:           []Sample{{Input: "5\n3 -1 2 2 0\n", Output: "-1 0 2 2 3\n"}},
		StarterCode:       "#include <algorithm>\n#include <iostream>\n#include <vector>\n\nint main() {\n    int n;\n    std::cin >> n;\n    std::vector<long long> values(n);\n    for (auto &value : values) std::cin >> value;\n    std::sort(values.begin(), values.end());\n    for (int i = 0; i < n; ++i) {\n        if (i) std::cout << ' ';\n        std::cout << values[i];\n    }\n    std::cout << '\\n';\n    return 0;\n}\n",
	},
	{
		ID: "guess-number", Type: "interactive", Title: "交互猜数字", Summary: "通过最多 10 次询问，在 1 到 1000 中找到评测器保存的秘密整数。", Difficulty: "入门·交互",
		Description:       "这是交互题。评测器保存一个 1 到 1000 的秘密整数，不会直接告诉你答案。你的程序需要与评测器对话，利用询问逐步确定答案。",
		InputDescription:  "评测器先输出一行 1000。每次收到你输出的 ? x 后，返回一个字符：< 表示秘密数小于 x，> 表示秘密数大于 x，= 表示秘密数等于 x。x 必须是 1 到 1000 的整数；最多询问 10 次。",
		OutputDescription: "询问时输出一行 ? x，找到答案后输出一行 ! x 并正常退出。每次输出后必须 flush（例如使用 std::endl）。正确答案判 AC；错误答案、非法格式、越界和超过 10 次询问判 WA。",
		Notes:             "每次输出后及时 flush；最终答案之后无需等待回复。",
		EngineeringNotes:  "固定协议 guess-number-v1。秘密数据只交给交互器，不作为选手标准输入。没有及时 flush 或等待不存在的回复可能导致 TLE。默认代码是可运行的二分查找演示解法。",
		Samples:           []Sample{{Input: "评测器：1000\n选手：? 500\n评测器：<\n选手：? 250\n评测器：=\n", Output: "选手：! 250\n（此对话只说明协议，实际秘密数未知）\n"}},
		StarterCode:       "#include <iostream>\n\nint main() {\n    int maximum;\n    if (!(std::cin >> maximum)) return 0;\n    int left = 1, right = maximum;\n    while (left <= right) {\n        int middle = left + (right - left) / 2;\n        std::cout << \"? \" << middle << std::endl;\n        char reply;\n        if (!(std::cin >> reply)) return 0;\n        if (reply == '=') {\n            std::cout << \"! \" << middle << std::endl;\n            return 0;\n        }\n        if (reply == '<') right = middle - 1;\n        else if (reply == '>') left = middle + 1;\n        else return 0;\n    }\n    return 0;\n}\n",
	},
	{
		ID: "construct-permutation", Title: "构造不相邻排列", Summary: "构造一个排列，使任意相邻的两个数之差不为 1。", Difficulty: "基础·构造",
		Description:       "给定整数 n，构造由 1 到 n 的每个整数恰好出现一次组成的排列，使任意相邻的两个整数的差的绝对值都不等于 1。可以输出任意一个满足条件的答案；若不存在，输出 -1。",
		InputDescription:  "输入一个整数 n（1 ≤ n ≤ 1000）。",
		OutputDescription: "若有解，输出 n 个整数，必须覆盖 1 到 n，不能重复或越界，且每对相邻数字之差的绝对值不能为 1。若无解，只输出 -1。不能输出其他内容。",
		Notes:             "答案可能不唯一，只要符合题目条件即可。n 为 1 时，一个数字本身就是合法排列。",
		EngineeringNotes:  "内置特殊判题 nonconsecutive-permutation-v1，读取 .in 并检查数学条件，不需要 .out。默认代码是可运行的演示解法。",
		Samples:           []Sample{{Input: "4\n", Output: "2 4 1 3\n"}, {Input: "3\n", Output: "-1\n"}},
		StarterCode:       "#include <iostream>\n\nint main() {\n    int n;\n    std::cin >> n;\n    if (n == 2 || n == 3) { std::cout << -1 << '\\n'; return 0; }\n    for (int x = 2; x <= n; x += 2) std::cout << x << ' ';\n    for (int x = 1; x <= n; x += 2) std::cout << x << ' ';\n    std::cout << '\\n';\n}\n",
	},
}

const communicationSumSender = "#include <iostream>\n\nint main() {\n    int n;\n    if (!(std::cin >> n)) return 1;\n    long long sum = 0, value;\n    for (int i = 0; i < n; ++i) {\n        if (!(std::cin >> value)) return 1;\n        sum += value;\n    }\n    std::cout << sum << '\\n';\n}\n"
const communicationMaximumSender = "#include <iostream>\n\nint main() {\n    int n;\n    long long answer, value;\n    if (!(std::cin >> n >> answer)) return 1;\n    for (int i = 1; i < n; ++i) {\n        if (!(std::cin >> value)) return 1;\n        if (value > answer) answer = value;\n    }\n    std::cout << answer << '\\n';\n}\n"
const communicationReceiver = "#include <iostream>\n\nint main() {\n    long long answer;\n    if (!(std::cin >> answer)) return 1;\n    std::cout << answer << '\\n';\n}\n"
const communicationPairSender = "#include <iostream>\n\nint main() {\n    long long a, b;\n    if (!(std::cin >> a >> b)) return 1;\n    std::cout << a << ' ' << b << '\\n';\n}\n"
const communicationPairReceiver = "#include <iostream>\n\nint main() {\n    long long a, b;\n    if (!(std::cin >> a >> b)) return 1;\n    std::cout << a + b << '\\n';\n}\n"

func init() {
	catalog = append(catalog, Definition{
		ID: "communication-sum", Title: "通信求和", Summary: "发送端读取数据，用有限通信帮助接收端输出整数和。", Difficulty: "基础·通信", Type: "communication",
		Description:       "提交发送端和接收端两份程序。只有发送端能读取原始数据；发送端的标准输出传给接收端的标准输入，接收端负责输出所有整数的和。两份程序均需正常退出。",
		InputDescription:  "发送端输入第一行整数 n（1 ≤ n ≤ 100），第二行 n 个整数，每个整数在 −10⁹ 到 10⁹ 之间。接收端只能读取发送端传来的消息。",
		OutputDescription: "接收端输出一个整数，表示所有输入整数的和。发送端向接收端发送的消息总量不能超过 32 字节。",
		Notes:             "发送端和接收端须提交为两个明确角色；两份程序自行约定消息格式，发送端完成输出后正常退出。",
		EngineeringNotes:  "固定单向通信图 sender stdout → receiver stdin；communication-v1 使用 token 比较接收端答案。32 字节包含消息中的空白和换行，按原始字节计数。每角色资源限额分别生效，共享墙钟截止时间。默认双模板是可运行演示解法。",
		Samples:           []Sample{{Input: "4\n1 -2 3 4\n", Output: "6\n"}}, StarterCode: communicationSumSender,
		StarterCodes: map[string]string{"sender": communicationSumSender, "receiver": communicationReceiver},
	}, Definition{
		ID: "communication-maximum", Title: "通信最大值", Summary: "发送端读取数组，用有限通信帮助接收端输出最大值。", Difficulty: "基础·通信", Type: "communication",
		Description:       "提交发送端和接收端两份程序。发送端读取数组，通过标准输出向接收端发送消息；接收端只能从自己的标准输入读取消息，并输出数组中的最大值。两份程序均需正常退出。",
		InputDescription:  "发送端输入第一行整数 n（1 ≤ n ≤ 100），第二行 n 个整数，每个整数在 −10⁹ 到 10⁹ 之间。原始数组不会直接提供给接收端。",
		OutputDescription: "接收端输出数组中的最大整数。发送端的消息总量不能超过 16 字节。",
		Notes:             "注意全为负数的数组。两份程序自行约定消息格式，发送端完成输出后正常退出。",
		EngineeringNotes:  "固定 sender stdout → receiver stdin 单向通信，原始消息最多 16 字节。接收端答案按 token 比较，两个容器独立资源限额，共享用例墙钟预算。默认双模板是可运行演示解法。",
		Samples:           []Sample{{Input: "4\n-8 -3 -10 -5\n", Output: "-3\n"}}, StarterCode: communicationMaximumSender,
		StarterCodes: map[string]string{"sender": communicationMaximumSender, "receiver": communicationReceiver},
	}, Definition{
		ID: "communication-a-plus-b", Title: "通信 A + B", Summary: "发送端传递两个整数，接收端计算并输出它们的和。", Difficulty: "入门·通信", Type: "communication",
		Description:       "提交发送端和接收端两份程序。发送端读取整数 a、b，通过标准输出把这两个整数传给接收端。接收端从自己的标准输入读取消息，计算 a+b 并输出答案。两份程序均需正常退出。",
		InputDescription:  "只有发送端能直接读取原始输入：一行两个整数 a、b，范围均为 −10¹² ≤ a, b ≤ 10¹²。接收端只能读取发送端传来的消息。",
		OutputDescription: "发送端传递 a 和 b，建议用空格分隔并在末尾换行。接收端输出一个整数，表示 a+b。每例发送端传递的消息总量不能超过 32 字节。",
		Notes:             "发送端负责传递原始两个数，接收端负责求和。使用 long long 保存输入和答案。发送端完成输出后正常退出，接收端读取两个数即可计算。",
		EngineeringNotes:  "固定单向通信图 sender stdout → receiver stdin；communication-v1 比较接收端答案，包含空格和换行在内的消息按原始字节计数。两角色各自受内存限制，共享用例墙钟预算。默认模板由发送端传递两个数、接收端求和；不由发送端预计算和。",
		Samples:           []Sample{{Input: "1 2\n", Output: "3\n"}}, StarterCode: communicationPairSender,
		StarterCodes: map[string]string{"sender": communicationPairSender, "receiver": communicationPairReceiver},
	})
}

func clone(p Definition) Definition {
	if p.Type == "" {
		p.Type = "batch"
	}
	p.Samples = append([]Sample(nil), p.Samples...)
	p.StarterCodes = CloneStarterCodes(p.StarterCodes)
	return p
}

func CloneStarterCodes(codes map[string]string) map[string]string {
	if codes == nil {
		return nil
	}
	result := make(map[string]string, len(codes))
	for role, source := range codes {
		result[role] = source
	}
	return result
}

// List returns owned copies in the display order. Callers cannot mutate the allowlist.
func List() []Definition {
	result := make([]Definition, len(catalog))
	for i, p := range catalog {
		result[i] = clone(p)
	}
	return result
}

func Get(id string) (Definition, bool) {
	for _, p := range catalog {
		if p.ID == id {
			return clone(p), true
		}
	}
	return Definition{}, false
}

func Exists(id string) bool {
	_, ok := Get(id)
	return ok
}
