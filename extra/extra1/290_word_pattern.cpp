/*
    leetcode 290 easy
    double hashmap one for s - pattern char
    one for patternchar - s
*/

class Solution {
public:
    bool wordPattern(string pattern, string s) {
        unordered_map<string,char> sToPattern;
        unordered_map<char,string> patternToS;
        stringstream ss(s);
        string tmp;
        int i = 0;
        while(ss >> tmp){
            if(patternToS.find(pattern[i])!=patternToS.end()){
                if(patternToS[pattern[i]] != tmp) return false;
            }
            if(sToPattern.find(tmp) != sToPattern.end()){
                if(sToPattern[tmp]!=pattern[i]) return false;
            }
            patternToS[pattern[i]] = tmp;
            sToPattern[tmp] = pattern[i];
            i++;
        }
        if(i != pattern.length()) return false;
        return true;
    }
};