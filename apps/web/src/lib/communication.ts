import type { CommunicationRole, CommunicationSources, SourceUnit } from './api'

export const communicationRoles: CommunicationRole[] = ['sender', 'receiver']
export const roleLabels: Record<CommunicationRole, string> = {
  sender: '发送端',
  receiver: '接收端',
}

export const communicationAdditionExample: CommunicationSources = {
  sender: `#include <iostream>
int main() {
    long long a, b;
    if (!(std::cin >> a >> b)) return 1;
    // Send both numbers; the receiver calculates their sum.
    std::cout << a << ' ' << b << '\\n';
    return 0;
}
`,
  receiver: `#include <iostream>
int main() {
    long long a, b;
    if (!(std::cin >> a >> b)) return 1;
    std::cout << a + b << '\\n';
    return 0;
}
`,
}

export const communicationAdditionSample = { input: '2 3\n', output: '5\n' }

export const communicationExample: CommunicationSources = {
  sender: `#include <iostream>
int main() {
    int n;
    if (!(std::cin >> n)) return 1;
    long long sum = 0, value;
    for (int i = 0; i < n; ++i) {
        if (!(std::cin >> value)) return 1;
        sum += value;
    }
    std::cout << sum << '\\n';
    return 0;
}
`,
  receiver: `#include <iostream>
int main() {
    long long sum;
    if (!(std::cin >> sum)) return 1;
    std::cout << sum << '\\n';
    return 0;
}
`,
}

export function sourceBytes(source: string) {
  return new TextEncoder().encode(source).byteLength
}

export function buildSourceUnits(sources: CommunicationSources): SourceUnit[] {
  const keys = Object.keys(sources)
  if (
    keys.length !== 2 ||
    keys.some((key) => !communicationRoles.includes(key as CommunicationRole))
  )
    throw new Error('必须提供发送端和接收端两份代码。')
  return communicationRoles.map((role) => {
    const source = sources[role]
    if (typeof source !== 'string' || !source.trim())
      throw new Error(`请输入${roleLabels[role]} C++ 代码。`)
    if (source.includes('\0'))
      throw new Error(`${roleLabels[role]}代码不能包含 NUL 字符。`)
    if (sourceBytes(source) > 65536)
      throw new Error(`${roleLabels[role]}代码超过 64 KiB，请精简。`)
    return { role, language: 'cpp', source }
  })
}

export function validateCommunicationLimit(value: string) {
  const limit = Number(value)
  if (!Number.isInteger(limit) || limit < 1 || limit > 65536)
    throw new Error('通信限额须为 1 至 65,536 字节的整数。')
  return limit
}

export function sourcesMatch(
  left: CommunicationSources,
  right: CommunicationSources,
) {
  return communicationRoles.every((role) => left[role] === right[role])
}
