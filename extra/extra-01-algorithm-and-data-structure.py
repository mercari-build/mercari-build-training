def word_pattern(p, s):
    s_tokens = s.split()
    patterns = {}  # key: characters, value: associated words 
    values = []   # keep track of values in patterns dictionary
    if len(s_tokens) != len(p):
        return False
    for token, char in zip(s_tokens, p):
        if not char in patterns:
            # the same token shouldn't be assigned to two different chars
            if token in values:  
                return False
            patterns[char] = token
            values.append(token)
        else:
            if patterns[char] != token:
                return False
    return True

def find_disappeared_numbers_1(nums):
    """
        with O(n^2)-time and O(1)-space
        :type nums: List[int]
        :rtype: List[int]
    """
    disappeared_nums = []
    n = len(nums)
    for i in range(1, n + 1):
        if i not in nums:
            disappeared_nums.append(i)
    return disappeared_nums

def find_disappeared_numbers_2(nums):
    """
        with with O(n)-time and O(n)-space
        :type nums: List[int]
        :rtype: List[int]
    """
    n = len(nums)
    # each index corresponds to num of index + 1; has true if there is the number
    # in the array and false otherwise
    appeared_nums = [False] * n  
    for num in nums:
        appeared_nums[num - 1] = True
    disappeared_nums = []
    for i in range(n):
        if appeared_nums[i] == False:
            disappeared_nums.append(i + 1)
    return disappeared_nums

def find_disappeared_numbers_3(nums):
    """
        with with O(n)-time and O(1)-space
        :type nums: List[int]
        :rtype: List[int]
    """
    # Again, each index corresponds to num of index + 1; instead of having another array,
    # keep track of whether num appears or not in the sign of the values; negative 
    # if it's in the array and false otherwise
    for num in nums:
        num = abs(num)
        if nums[num - 1] > 0:
            nums[num - 1] *= -1   # nagate the num
    disappeared_nums = []
    for i in range(len(nums)):
        if nums[i] > 0:
            disappeared_nums.append(i + 1)
    return disappeared_nums


# Definition for singly-linked list.
class ListNode(object):
    def __init__(self, x):
        self.val = x
        self.next = None

def get_intersection_node_1(headA, headB):
    """
    with O(n)-time and O(n)-space
    :type head1, head1: ListNode
    :rtype: ListNode
    """
    nodes = set()
    while headA:
        nodes.add(headA)
        headA = headA.next
    while headB:
        if headB in nodes:
            return headB
        headB = headB.next
    return None

def count(head):
    """
    :type head: ListNode
    :rtype: int (the length of linked list)
    """
    num = 0
    while head != None:
        num += 1
        head = head.next
    return num

def get_intersection_node_2(headA, headB):
    """
    with O(n)-time and O(1)-space
    :type head1, head1: ListNode
    :rtype: ListNode
    """
    # Compare the length of two linked lists
    diff = count(headA) - count(headB)

    # Adjust the lengths 
    if diff > 0:  # head A is longer
        for i in range(diff):
            headA = headA.next
    else:
        for i in range(abs(diff)):
            headB = headB.next

    while headA and headB:
        if headA == headB: return headA
        headA = headA.next
        headB = headB.next
    return None


#######################################################################
"""
Tests
"""
def test_word_pattern():
    p1 = "abba"
    s1 = "dog cat cat dog"
    assert(word_pattern(p1, s1) == True)

    p2 = "abba"
    s2 = "dog cat cat fish"
    assert(word_pattern(p2, s2) == False)

    p3 = "aaaa"
    s3 = "apple apple apple apple"
    assert(word_pattern(p3, s3) == True)

    p4 = "aba"
    s4 = "apple banana apple apple"
    assert(word_pattern(p4, s4) == False)

    p5 = "a"
    s5 = "dog"
    assert(word_pattern(p5, s5) == True)

    p6 = "abba"
    s6 = "dog dog dog dog"
    assert(word_pattern(p6, s6) == False)

    print("all word pattern tests passed!")


def test_find_disappeared_numbers():
    nums1 = [4, 3, 2, 7, 8, 2, 3, 1]
    assert(find_disappeared_numbers_1(nums1) == [5, 6])
    assert(find_disappeared_numbers_2(nums1) == [5, 6])
    assert(find_disappeared_numbers_3(nums1) == [5, 6])

    nums2 = [1, 2, 3, 4, 5]
    assert(find_disappeared_numbers_1(nums2) == [])
    assert(find_disappeared_numbers_2(nums2) == [])
    assert(find_disappeared_numbers_3(nums2) == [])

    nums3 = [5, 4, 3, 2, 1]
    assert(find_disappeared_numbers_1(nums3) == [])
    assert(find_disappeared_numbers_2(nums3) == [])
    assert(find_disappeared_numbers_3(nums3) == [])

    nums4 = [1, 1, 1, 1, 1]
    assert(find_disappeared_numbers_1(nums4) == [2, 3, 4, 5])
    assert(find_disappeared_numbers_2(nums4) == [2, 3, 4, 5])
    assert(find_disappeared_numbers_3(nums4) == [2, 3, 4, 5])

    nums5 = [1]
    assert(find_disappeared_numbers_1(nums5) == [])
    assert(find_disappeared_numbers_2(nums5) == [])
    assert(find_disappeared_numbers_3(nums5) == [])

    nums6 = [1, 1]
    assert(find_disappeared_numbers_1(nums6) == [2])
    assert(find_disappeared_numbers_2(nums6) == [2])
    assert(find_disappeared_numbers_3(nums6) == [2])

    print("all find_disappeared_numbers tests passed!")

#######################################################################


def main():
    test_word_pattern()
    test_find_disappeared_numbers()
    # get_intersection_node() tested in leetcode 160


if __name__== "__main__":
    main()