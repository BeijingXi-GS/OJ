import type { BatchChecker, ProblemType, UploadDescriptor } from './api'

export const constructionRule =
  '输入一个整数 n（1 ≤ n ≤ 1000）。输出 1 到 n 的一个排列，任意相邻两个数的差的绝对值不能为 1。n = 2 或 3 时只输出 -1；其余情况必须给出合法排列。允许不同的正确排列，不得重复、遗漏、越界或输出多余内容。'

// 下载与发布均使用这份源文本，避免示例和实际提交的程序不一致。
export const checkerExample = `// C++20 special judge example; no third-party library is required.
// argv[1] input, argv[2] optional reference answer, argv[3] contestant output.
// exit 0 = accepted; exit 1 = wrong answer; every other exit = judge error.
#include <fstream>
#include <iostream>
#include <string>
#include <vector>

static int reject(const char* message) {
    std::cerr << message << '\\n';
    return 1;
}

static bool decimal(const std::string& token, int& value) {
    if (token.empty()) return false;
    value = 0;
    for (char digit : token) {
        if (digit < '0' || digit > '9') return false;
        value = value * 10 + digit - '0';
        if (value > 1000) return false;
    }
    return true;
}

int main(int argc, char** argv) {
    if (argc != 4) { std::cerr << "Checker requires three file arguments.\\n"; return 2; }
    std::ifstream input(argv[1]), answer(argv[2]), output(argv[3]);
    if (!input || !answer || !output) { std::cerr << "Cannot open judge files.\\n"; return 2; }
    std::string token, extra;
    int n;
    if (!(input >> token) || !decimal(token, n) || n < 1 || n > 1000 || (input >> extra)) {
        std::cerr << "Invalid problem input.\\n";
        return 2;
    }
    // This construction condition does not need a reference answer.
    if (n == 2 || n == 3) {
        if (!(output >> token) || token != "-1" || (output >> extra)) return reject("No solution: output only -1.");
        if (output.bad()) { std::cerr << "Output file read failed.\\n"; return 2; }
        std::cerr << "Correctly reported no solution.\\n";
        return 0;
    }
    std::vector<bool> seen(n + 1, false);
    int previous = 0;
    for (int i = 0; i < n; ++i) {
        int value;
        if (!(output >> token)) return reject("Too few integers.");
        if (!decimal(token, value) || value < 1 || value > n) return reject("Every value must be an integer from 1 to n.");
        if (seen[value]) return reject("Repeated value in permutation.");
        if (i && (previous - value == 1 || value - previous == 1)) return reject("Adjacent values differ by one.");
        seen[value] = true;
        previous = value;
    }
    if (output >> extra) return reject("Extra output after the permutation.");
    if (output.bad()) { std::cerr << "Output file read failed.\\n"; return 2; }
    std::cerr << "All construction conditions are satisfied.\\n";
    return 0;
}
`

export function buildCases(
  files: Pick<File, 'name' | 'size'>[],
  type: ProblemType,
  checker: BatchChecker,
): UploadDescriptor['cases'] {
  if (files.length === 0) throw new Error('请选择至少一组测试数据。')
  if (files.length > 16) throw new Error('最多上传 16 个数据文件。')
  const names = new Set<string>()
  for (const file of files) {
    const match = /^([a-z0-9][a-z0-9_-]{0,63})\.(in|out)$/.exec(file.name)
    if (!match || match[0] !== file.name)
      throw new Error(
        `文件 ${file.name} 命名不符合要求，请以小写字母或数字开头，其后可用小写字母、数字、下划线或短横线（主体最多 64 字符），加 .in / .out 后缀。`,
      )
    if (/^(con|prn|aux|nul)$|^(com|lpt)/.test(match[1]))
      throw new Error(
        `文件 ${file.name} 使用了保留名称；请避开 con、prn、aux、nul 及 com / lpt 开头的主体，例如改为 01.in。`,
      )
    if (names.has(file.name)) throw new Error(`文件 ${file.name} 重复。`)
    if (file.size > 65536) throw new Error(`文件 ${file.name} 超过 64 KiB。`)
    names.add(file.name)
  }
  const inputs = files
    .filter((file) => file.name.endsWith('.in'))
    .sort((a, b) => a.name.localeCompare(b.name))
  const custom = type === 'batch' && checker === 'cpp-special-v1'
  const onlyInput =
    type === 'interactive' ||
    (type === 'batch' && checker === 'nonconsecutive-permutation-v1')
  const maximum = custom || type === 'communication' ? 4 : 8
  if (inputs.length === 0 || inputs.length > maximum)
    throw new Error(`测试数据须包含 1 至 ${maximum} 个 .in 文件。`)
  const cases = inputs.map((file, index) => {
    const expected = `${file.name.slice(0, -3)}.out`
    if (!onlyInput && !custom && !names.has(expected))
      throw new Error(`缺少与 ${file.name} 配对的 ${expected}。`)
    return {
      id: `case-${String(index + 1).padStart(2, '0')}`,
      input: file.name,
      ...(!onlyInput && names.has(expected) ? { expected } : {}),
    }
  })
  const referenced = new Set(
    cases.flatMap((item) =>
      item.expected ? [item.input, item.expected] : [item.input],
    ),
  )
  if (files.some((file) => !referenced.has(file.name)))
    throw new Error(
      onlyInput
        ? '该判定方式只需 .in 测试输入，不上传 .out。'
        : '存在没有对应 .in 的 .out 文件，请检查配对。',
    )
  return cases
}

export function validateCheckerSource(source: string) {
  if (!source.trim()) throw new Error('请填写或选择 C++ 特判源码。')
  if (source.includes('\0')) throw new Error('特判源码不能包含 NUL 字符。')
  if (new TextEncoder().encode(source).byteLength > 65536)
    throw new Error('特判源码超过 64 KiB，请精简。')
}

export function isConstructionInput(input: string) {
  const match = /^[\t\n\v\f\r ]*([0-9]{1,4})[\t\n\v\f\r ]*$/.exec(input)
  return !!match && Number(match[1]) >= 1 && Number(match[1]) <= 1000
}

export function isInteractiveInput(input: string) {
  const match = /^[\t\n\r ]*([0-9]{1,4})[\t\n\r ]*$/.exec(input)
  return !!match && Number(match[1]) >= 1 && Number(match[1]) <= 1000
}
