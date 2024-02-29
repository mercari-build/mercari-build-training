package main

import "fmt"

func findDisappearedNumbers(nums []int) []int {
    n := len(nums)
    numMap := make(map[int]bool)
    result := []int{}

    for _, num := range nums {
        numMap[num] = true
    }

    for i := 1; i <= n; i++ {
        if !numMap[i] {
            result = append(result, i)
        }
    }

    return result
}

func findDisappearedNumbers2(nums []int) []int {
    for _, num := range nums {
        index := abs(num) - 1
        if nums[index] > 0 {
            nums[index] = -nums[index]
        }
    }

    result := []int{}
    for i, num := range nums {
        if num > 0 {
            result = append(result, i+1)
        }
    }

    return result
}

// Helper function to get absolute value
func abs(a int) int {
    if a < 0 {
        return -a
    }
    return a
}

func findDisappearedNumbers3(nums []int) []int {
    appeared := make([]bool, len(nums))
    result := []int{}

    for _, num := range nums {
        appeared[num-1] = true
    }

    for i, appeared := range appeared {
        if !appeared {
            result = append(result, i+1)
        }
    }

    return result
}
