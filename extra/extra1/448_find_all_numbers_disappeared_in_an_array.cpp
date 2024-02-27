/*
    leetcode 448 easy
    use vector<bool> map, vector<ans>
    turn true to every nums[i]-1
    loop again and push back in ans such that map[i] still has false 
*/

class Solution {
public:
    vector<int> findDisappearedNumbers(vector<int>& nums) {
        int n = nums.size();
        vector<bool> map (n,false);
        vector<int> ans;
        for(int i = 0; i<n; i++){
            map[nums[i]-1] = true;
        }
        for(int i = 0; i<n; i++){
            if(map[i] == false) ans.push_back(i+1);
        }
        return ans;
    }
};