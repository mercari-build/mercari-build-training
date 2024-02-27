class Solution:
    def lengthOfLongestSubstring(self, s: str) -> int:
        tmp_ans = ""
        ans = ""
        d = {chr(i): False for i in range(ord('a'), ord('z')+1)}
        for char in s:
            if not d[char]:
                tmp_ans += char
            else:
                if len(ans) < len(tmp_ans):
                    ans = tmp_ans
                tmp_ans = char
                d = {chr(i): False for i in range(ord('a'), ord('z')+1)}
            d[char] = True
        return len(ans)