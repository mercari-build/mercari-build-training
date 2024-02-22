class Solution:
    def wordPattern(self, pattern: str, s: str) -> bool:
        s_list = s.split()

        if len(pattern) != len(s_list):
            return False

        d = {}
        for pattern_char, s_char in zip(pattern,s_list):
            if pattern_char not in d:
                d[pattern_char] = s_char
            else:
                if d[pattern_char] != s_char:
                    return False
        return True


        