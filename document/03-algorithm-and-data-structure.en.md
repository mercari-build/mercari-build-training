# STEP3: Algorithms and Data Structures

In STEP3, after learning about basic algorithms and data structures, you will practice using problems from LeetCode. It is recommended to solve the problems before looking at the explanations.

## Textbooks

**:book: Reference**

* (JA) [Algorithms, Data Structures, and Coding Test Introduction Taught by a Silicon Valley Engineer](https://mercari.udemy.com/course/python-algo/)

* (EN) [Python Data Structures & Algorithms + LEETCODE Exercises](https://mercari.udemy.com/course/data-structures-algorithms-python/)

**:beginner: Point**
* First, look up the following terms and be able to explain what they are.
    * Time complexity and space complexity:
        Time complexity: measured in the number of operations it take to complete a task.

        Space complexity: measured in the amount of memory space a code takes up.

    * Big O notation: a way of comparing two sets of code to determine which is more efficient.
        Big O (from most to least efficient):
            O(1) = constant time
            O(log n) = divide and conquer
            O(n) = proportional
            O(n^2) = a loop within a loop

    * Associative arrays: a data structure that stores key-value pairs, known as a dictionary (JS = Object)


* Study the following basic algorithms on Udemy or other platforms and be able to explain them.
    * What is binary search? Explain why the time complexity of binary search is $O(\log n)$.
    A binary search is an efficient way to find a target within an assorted array. In this algorithm, the array is split in half. The half containing the target will be split again. This will repeat until the target is found.

    * Explain the difference between LinkedList and Array.
    LinkedList:
Stored dynamically (no index)

- Append O(1)
- Pop O(n)
- Prepend O(1) = more efficient than an Array
- Pop First O(1) = more efficient than an Array
- Insert O(n)
- Remove O(n)
- Lookup by Index O(n)
- Lookup by Value O(n)
        

    Array(List):
Fixed positioned (index can be used)

Append O(1)
- Pop O(1) = more efficient than a LinkedList
- Prepend O(n)
- Pop First O(n)
Insert O(n)
- Remove O(n)
- Lookup by Index O(1) = more efficient than a LinkedList
- Lookup by Value O(n)

    * Explain the hash table and estimate the time complexity.
    * Explain graph search algorithms and explain the difference between BFS (Breadth First Search) and DFS (Depth First Search).


## Exercises
### [Word Pattern](https://leetcode.com/problems/word-pattern/description/)
Given a pattern `p` consisting of lowercase English letters and a string `s` separated by spaces, determine if `s` follows the pattern `p`. For example, if `p = "abba"` and `s = "dog cat cat dog"`, then `s` follows the pattern `p`, but if `p="abba"` and `s="dog cat cat fish"`, then `s` does not follow `p`.
sudo_code:
##split s
##create dictionary {}
##if p is new key=p.value, value=s.value, else compare
##if s.value != {p.value:value}, return false
##return true

**:beginner: Checkpoint**
#### Step1: Think about how to split the string `s` by spaces.
<details>
<summary>Hint</summary>

* In each language, there should be standard libraries or functions provided for string manipulation.
* Use web search or ChatGPT, searching for "split string by spaces" or similar queries.
</details>

class Solution(object):
    def wordPattern(self, pattern, s):
        word = s.split(" ")

#### Step2: Consider how to manage which part of `s` corresponds to each character of the pattern `p`.
<details>
<summary>Hint</summary>

* For example, in Example 1, the words in `s` corresponding to each character of `p` are `a => dog`, `b => cat`.
* To manage such correspondences, using a dictionary or hash table would be beneficial.
* For instance, in Python, you can manage the words in `s` that correspond to each character of `p` using a `dict`.
* Also use web search or ChatGPT, looking up "Python dictionary" or similar queries.
</details>

class Solution(object):
    def wordPattern(self, pattern, s):
        word = s.split(" ")

##added the length check after failing submission
         if len(pattern) != len(words):
            return False

        pattern_dictionary = {}
        word_dictionary = {}

        for i in range(len(pattern)):
            char = pattern[i]

            if char in pattern_dictionary:
                if pattern_dictionary[char]!=word[i]:
                    return False
            else:
                if word[i] in word_dictionary:
                    return False
                else:
                    pattern_dictionary[char] = word[i]
                    word_dictionary[word[i]] = char
        
        return True


### [Find All Numbers Disappeared in an Array](https://leetcode.com/problems/find-all-numbers-disappeared-in-an-array/description/)
Given an array of n integers where each value is in the range [1, n], return all integers in the range [1, n] that do not appear in the array.
sudo_code:
##determine length
##create empty list []
##iterate through range of the sorted list to find missing num
##push missing num to []
##return list

**:beginner: Checkpoint**

#### Step1: Solve with O(n^2)-time and O(1)-space
<details>
<summary>Hint</summary>
* You can solve it using a simple double loop, achieving O(n^2)-time and O(1)-space.
</details>
class Solution(object):
    def findDisappearedNumbers(self, nums):
        allNums = []

        for i in range(len(nums)):
            if nums[i] > 0:
                allNums.append(i + 1)
         
        return([num for num in allNums if num not in nums])
##Test succeeded but time limit exceeded for Case 3

#### Step2: Solve with O(n)-time and O(n)-space
<details>
<summary>Hint</summary>

* By preparing an array to record whether an element has appeared in the array nums, you can solve it in O(n)-time and O(n)-space.
</details>

class Solution(object):
    def findDisappearedNumbers(self, nums):
        results = [] 

##added abs(i) - 1 to make nums negative (method to keep track of marked nums without increasing space complexity) 
        for i in nums:
            pos = abs(i) - 1
            if nums[pos] > 0:
                nums[pos] *= -1

        for i in range(len(nums)):
            if nums[i] > 0:
                results.append(i + 1)

        return results

#### Advanced: Solve with O(n)-time and O(1)-space (bonus)
Is it possible to solve it in O(1)-space, excluding the input and return?
<details>
<summary>Hint</summary>

* Upon deeper consideration, it is possible to solve it in O(n)-time and O(1)-space.
* This will be covered in the explanation, so give it a try.
</details>


### [Intersection of Two Linked Lists](https://leetcode.com/problems/intersection-of-two-linked-lists/description)
Given two singly linked lists, return the node at which the two lists intersect. If the two linked lists have no intersection at all, return `null`.
##create a node for list1, list2
##set pointers at the heads of l1, l2
##iterate though both lists
##when l1(head) = l2(head) return node

**:beginner: Checkpoint**

#### Step1: Solve with O(n)-time and O(n)-space
<details>
<summary>Hint</summary>

* By using a Hash Table to record nodes, you can solve it in O(n)-time and O(n)-space.
</details>

class Solution(object):
    def getIntersectionNode(self, headA, headB):
       listA = headA
       listB = headB
       while listA != listB:
        listA = listA.next if listA else headB
        listB = listB.next if listB else headA
       return listA

#### Step2: Solve with O(n)-time and O(1)-space
Is it possible to solve it in O(1)-space, excluding the input and return?
<details>
<summary>Hint</summary>

* By comparing the lengths of the two lists and adjusting the longer list to match the length of the shorter one, you can solve it in O(n)-time and O(1)-space.
* This will be covered in the explanation.
</details>

#### Advanced: Solve using two pointers (bonus)
<details>
<summary>Hint</summary>

* Start a pointer at the tail of one list and proceed to the head, reducing the problem to Floyd's Linked List Cycle Finding Algorithm.
</details>


### [Koko Eating Bananas](https://leetcode.com/problems/koko-eating-bananas/) (optional)

### [Non-overlapping Intervals](https://leetcode.com/problems/non-overlapping-intervals/description/) (optional)

### [Longest Substring Without Repeating Characters](https://leetcode.com/problems/longest-substring-without-repeating-characters/description/) (optional)

[STEP4: Make a listing API](./04-api.en.md)