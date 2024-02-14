# Extra1: Algorithms and Data Structures

In Extra1, after learning the basic algorithms and data structures, we will practice using problems from LeetCode. On Thursday, 2/22, from 18:00, we will provide problem explanations in Japanese. It is recommended to solve the problems in advance.

## Textbooks

**:book: Reference**

* (JA) [Algorithms, Data Structures, and Coding Test Introduction Taught by a Silicon Valley Engineer](https://mercari.udemy.com/course/python-algo/)

* (EN) [Python Data Structures & Algorithms + LEETCODE Exercises](https://mercari.udemy.com/course/data-structures-algorithms-python/)

**:beginner: Point**
* Please explain time complexity and space complexity.
* Please explain Big O notation.
* Please explain associative arrays.
* What is a binary search, and why is its computational complexity \(O(\log n)\)?
* Explain the difference between LinkedList and Array.
* Explain hash tables and estimate their computational complexity.
* Explain graph search algorithms and discuss the use cases for BFS/DFS.

## Exercises
### [Word Pattern](https://leetcode.com/problems/word-pattern/description/)
Given a pattern `p` consisting of lowercase English letters and a string `s` separated by spaces, determine if `s` follows the pattern `p`. For example, if `p = "abba"` and `s = "dog cat cat dog"`, then `s` follows the pattern `p`, but if `p="abba"` and `s="dog cat cat fish"`, then `s` does not follow `p`.

**:beginner: Checkpoint**
#### Step1: Think about how to split the string `s` by spaces.
<details>
<summary>Hint</summary>

* In each language, there should be standard libraries or functions provided for string manipulation.
* Use web search or ChatGPT, searching for "split string by spaces" or similar queries.
</details>

#### Step2: Consider how to manage which part of `s` corresponds to each character of the pattern `p`.
<details>
<summary>Hint</summary>

* For example, in Example 1, the words in `s` corresponding to each character of `p` are `a => dog`, `b => cat`.
* To manage such correspondences, using a dictionary or hash table would be beneficial.
* For instance, in Python, you can manage the words in `s` that correspond to each character of `p` using a `dict`.
* Also use web search or ChatGPT, looking up "Python dictionary" or similar queries.
</details>


### [Find All Numbers Disappeared in an Array](https://leetcode.com/problems/find-all-numbers-disappeared-in-an-array/description/)
Given an array of n integers where each value is in the range [1, n], return all integers in the range [1, n] that do not appear in the array.

**:beginner: Checkpoint**

#### Step1: Solve with O(n^2)-time and O(1)-space
<details>
<summary>Hint</summary>

* You can solve it using a simple double loop, achieving O(n^2)-time and O(1)-space.
</details>

#### Step2: Solve with O(n)-time and O(n)-space
<details>
<summary>Hint</summary>

* By preparing an array to record whether an element has appeared in the array nums, you can solve it in O(n)-time and O(n)-space.
</details>

#### Advanced: Solve with O(n)-time and O(1)-space (bonus)
Is it possible to solve it in O(1)-space, excluding the input and return?
<details>
<summary>Hint</summary>

* Upon deeper consideration, it is possible to solve it in O(n)-time and O(1)-space.
* This will be covered in the explanation, so give it a try.
</details>


### [Intersection of Two Linked Lists](https://leetcode.com/problems/intersection-of-two-linked-lists/description)
Given two singly linked lists, return the node at which the two lists intersect. If the two linked lists have no intersection at all, return `null`.

**:beginner: Checkpoint**

#### Step1: Solve with O(n)-time and O(n)-space
<details>
<summary>Hint</summary>

* By using a Hash Table to record nodes, you can solve it in O(n)-time and O(n)-space.
</details>

#### Step2: Solve with O(n)-time and O(1)-space
Is it possible to solve it in O(1)-space, excluding the input and return?
<details>
<summary>Hint</summary>

* By comparing the lengths of the two lists and adjusting the longer list to match the length of the shorter one, you can solve it in O(n)-time and O(1