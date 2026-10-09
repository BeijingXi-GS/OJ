// C++20 special judge example; no third-party library is required.
// argv[1] input, argv[2] optional reference answer, argv[3] contestant output.
// exit 0 = accepted; exit 1 = wrong answer; every other exit = judge error.
#include <fstream>
#include <iostream>
#include <string>
#include <vector>

static int reject(const char* message) {
    std::cerr << message << '\n';
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
    if (argc != 4) { std::cerr << "Checker requires three file arguments.\n"; return 2; }
    std::ifstream input(argv[1]), answer(argv[2]), output(argv[3]);
    if (!input || !answer || !output) { std::cerr << "Cannot open judge files.\n"; return 2; }
    std::string token, extra;
    int n;
    if (!(input >> token) || !decimal(token, n) || n < 1 || n > 1000 || (input >> extra)) {
        std::cerr << "Invalid problem input.\n";
        return 2;
    }
    // This construction condition does not need a reference answer.
    if (n == 2 || n == 3) {
        if (!(output >> token) || token != "-1" || (output >> extra)) return reject("No solution: output only -1.");
        if (output.bad()) { std::cerr << "Output file read failed.\n"; return 2; }
        std::cerr << "Correctly reported no solution.\n";
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
    if (output.bad()) { std::cerr << "Output file read failed.\n"; return 2; }
    std::cerr << "All construction conditions are satisfied.\n";
    return 0;
}
