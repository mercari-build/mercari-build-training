class Solution {
    public:
        vector<int> findDisappearedNumbers(vector<int>& nums) {
            vector<int> ans;
            vector<int> numberExists(nums.size() + 1, 0);
            for(int i=0;i<nums.size();i++){
                numberExists[nums[i]]++;
            }
            for(int i=1;i<numberExists.size();i++){
                if(numberExists[i]==0){
                    ans.push_back(i);
                }
            }
    
            return ans;    
    
        }
    };
