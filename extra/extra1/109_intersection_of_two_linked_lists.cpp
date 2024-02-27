/*
    leetcode 160
    iterate through node A first, and then
    go through B. record A with unordered_set<ListNode*>
    and when B is not found in set, return b.
*/
/**
 * Definition for singly-linked list.
 * struct ListNode {
 *     int val;
 *     ListNode *next;
 *     ListNode(int x) : val(x), next(NULL) {}
 * };
 */
class Solution {
public:
    ListNode *getIntersectionNode(ListNode *headA, ListNode *headB) {
        unordered_set<ListNode*> set;
        while(headA != nullptr){
            set.insert(headA);
            headA = headA->next;
        }
        while(headB != nullptr){
            if(set.find(headB)!=set.end()) return headB;
            headB = headB->next;
        }
        return nullptr;
    }
};