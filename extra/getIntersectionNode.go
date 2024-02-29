func getIntersectionNode(headA, headB *ListNode) *ListNode {
    lenA, lenB := 0, 0
    for node := headA; node != nil; node = node.Next {
        lenA++
    }
    for node := headB; node != nil; node = node.Next {
        lenB++
    }

    // Adjust starting points
    for lenA < lenB {
        headB = headB.Next
        lenB--
    }
    for lenB < lenA {
        headA = headA.Next
        lenA--
    }

    // Find intersection
    for headA != nil && headB != nil {
        if headA == headB {
            return headA
        }
        headA = headA.Next
        headB = headB.Next
    }
    return nil
}

func getIntersectionNode2(headA, headB *ListNode) *ListNode {
    // Create cycle in the first list
    originalTail := headA
    for originalTail != nil && originalTail.Next != nil {
        originalTail = originalTail.Next
    }
    // Edge case: one of the lists is empty
    if originalTail == nil {
        return nil
    }
    originalTail.Next = headA // Create cycle

    // Use Floyd's cycle finding algorithm on the second list
    slow, fast := headB, headB
    for fast != nil && fast.Next != nil {
        slow = slow.Next
        fast = fast.Next.Next
        if slow == fast { // Cycle detected
            slow = headB
            for slow != fast {
                slow = slow.Next
                fast = fast.Next
            }
            originalTail.Next = nil // Remove cycle
            return slow // Intersection node
        }
    }

    originalTail.Next = nil // Ensure to remove cycle if no intersection found
    return nil
}
