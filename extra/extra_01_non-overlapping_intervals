class Solution:
    def eraseOverlapIntervals(self, intervals: List[List[int]]) -> int:
        ans = 0
        pre_end = intervals[0][1]
        for i in range(1, len(intervals)):
            start = intervals[i][0]
            if pre_end > start:
                ans += 1
            pre_end = intervals[i][1]
        return ans