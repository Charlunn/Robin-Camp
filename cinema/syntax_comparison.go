package main // 1. 程序结构: 包声明
// 在 Go 中，每个文件都必须属于一个包。`package main` 表示这是一个可独立执行的程序，而不是一个库。
// 这类似于 C 需要一个包含 `main` 函数的文件来创建可执行文件。

// 2. 引入库 (Import)
import (
	"errors"
	"fmt"
)
// `import` 语句用于引入其他包。它类似于 C 的 `#include <stdio.h>`。
// Go 的标准库被分成了很多包，`fmt` 用于格式化 I/O (类似 stdio)，`errors` 用于创建错误对象。
// Go 会严格检查：引入了但未使用的包会导致编译错误。

// --- 函数定义 ---
/*
	// C 语言: 返回类型 函数名(参数类型 参数名, ...)
	int add(int a, int b) {
		return a + b;
	}
*/
// Go 语言: func 函数名(参数名 参数类型, ...) (返回类型1, 返回类型2, ...)
// Go 的函数可以返回多个值，这是其错误处理的关键特性。
func divide(a, b int) (int, error) { // `error` 是一个内置的接口类型
	if b == 0 {
		// 返回一个零值和一个错误信息
		return 0, errors.New("除数不能为零")
	}
	// 返回计算结果和 nil (代表没有错误)
	return a / b, nil
}

// --- 结构体与方法 ---
/*
	// C 语言: 结构体只包含数据
	struct Rect {
		double width;
		double height;
	};
	// 功能函数与数据是分离的
	double area(struct Rect r) {
		return r.width * r.height;
	}
*/
// Go 语言: 可以为结构体绑定"方法"，这让 Go 具有面向对象编程的风格。
// 9. 可见性规则: 首字母大写的名称 (如 Point, X, Y, Area) 是“导出”的，即公共的(Public)，可以在包外访问。
// 首字母小写的名称是“未导出”的，即私有的(private)，只能在包内访问。
type Rect struct {
	Width  float64 // 公共字段
	Height float64 // 公共字段
}

// `(r Rect)` 被称为“接收者(receiver)”，它将 `Area` 函数绑定到了 `Rect` 结构体上。
// 现在 `Area` 是 `Rect` 的一个方法，而不是一个独立的函数。
func (r Rect) Area() float64 {
	return r.Width * r.Height
}

// 3. 主函数 (程序入口)
/*
	// C 语言: int main(int argc, char *argv[])
	int main() {
		// ...
		return 0;
	}
*/
// Go 语言: func main()
// Go 的 main 函数没有参数，也没有返回值。命令行参数通过 `os` 包来获取。
func main() {
	// --- 4. 变量声明与数据类型 ---
	/*
		// C 语言: 类型 变量名 = 值;
		int a = 10;
		float b = 3.14;
		const double PI = 3.14159;
		char* s = "hello";
	*/
	// Go 语言:
	// a) 完整声明: var 变量名 类型 = 值
	var a int = 10
	// b) 类型推断: var 变量名 = 值
	var b = 3.14 // Go 会自动推断为 float64 类型
	// c) 短声明 (最常用): 变量名 := 值 (只能在函数内部使用)
	s := "hello" // Go 会自动推断为 string 类型

	// Go 有内置的 `bool` 类型
	var isGoCool bool = true

	// 常量 `const`
	const PI = 3.14159

	// Go 有“零值”概念。如果只声明不初始化，变量会自动获得其类型的零值。
	// int 的零值是 0, float 的是 0.0, bool 的是 false, string 的是 ""
	var zeroInt int
	fmt.Printf("变量: a=%d, b=%.2f, s=%s, isGoCool=%t, PI=%.5f, zeroInt=%d\n", a, b, s, isGoCool, PI, zeroInt)

	// --- 5. 指针 ---
	/*
		// C 语言: 强大的指针，支持指针运算
		int c_val = 100;
		int* c_ptr = &c_val;
		*c_ptr = 200;
		// c_ptr++; // 指针算术，移动到下一个内存地址
	*/
	// Go 语言: 更安全的指针
	go_val := 100
	var go_ptr *int = &go_val // 声明和 C 类似，& 获取地址

	fmt.Printf("\n--- 指针 ---\n原始值: %d, 指针地址: %p, 指针解引用: %d\n", go_val, go_ptr, *go_ptr)
	*go_ptr = 200 // * 解引用来修改值，和 C 一样
	fmt.Printf("修改后的值: %d\n", go_val)
	// go_ptr++ // !!! 编译错误 !!! Go 为了内存安全，不支持指针算术运算。
	// Go 的空指针是 `nil`，等同于 C 的 `NULL`。
	var nil_ptr *int
	fmt.Printf("空指针: %v\n", nil_ptr)

	// --- 6. 控制流 ---
	fmt.Println("\n--- 控制流 ---")
	// a) if-else: 无括号，支持初始化语句
	if num := 9; num < 0 {
		fmt.Println(num, "是负数")
	} else if num < 10 {
		fmt.Println(num, "是一位数") // `num` 的作用域只在 if-else 块内
	} else {
		fmt.Println(num, "是多位数")
	}

	// b) for: Go 唯一的循环关键字
	// C 风格的 for
	for i := 0; i < 3; i++ {
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// c) switch: Go 的 switch 更强大
	// 默认有 `break` (不会自动“掉落”到下一个 case)，case 可以是表达式
	grade := "B"
	switch grade {
	case "A":
		fmt.Println("优秀!")
	case "B", "C":
		fmt.Println("良好")
	default:
		fmt.Println("及格")
	}

	// --- 7. 复合类型 ---
	fmt.Println("\n--- 复合类型 ---")
	// a) 数组 (Array): 固定长度，值类型
	var arr [3]int = [3]int{1, 2, 3}

	// b) 切片 (Slice): 动态长度，引用类型 (更常用)
	// 你可以把它看作 C++ 的 vector 或 Python 的 list
	slice := []int{10, 20, 30}               // 创建一个切片
	slice = append(slice, 40)                 // 动态添加元素
	fmt.Printf("数组: %v, 切片: %v\n", arr, slice)

	// for...range 循环是遍历数组、切片、map 的最佳方式
	for index, value := range slice {
		fmt.Printf("切片索引: %d, 值: %d\n", index, value)
	}

	// c) Map (哈希表): Go 内置的键值对集合
	// 类似于 C++ 的 std::map 或 Python 的 dict
	ages := make(map[string]int) // 创建一个 map
	ages["Alice"] = 30
	ages["Bob"] = 25
	fmt.Printf("Map: %v\n", ages)
	delete(ages, "Bob") // 删除键
	// 检查键是否存在
	val, ok := ages["Bob"]
	if !ok {
		fmt.Println("Bob 的年龄不存在")
	} else {
		fmt.Println("Bob 的年龄是", val)
	}

	// --- 8. 错误处理 ---
	fmt.Println("\n--- 错误处理 ---")
	// Go 通过函数的多返回值来处理错误，这是 Go 的核心模式
	result, err := divide(10, 2)
	if err != nil { // 如果 err 不是 nil，说明有错误发生
		fmt.Println("错误:", err)
	} else {
		fmt.Println("10/2 的结果是:", result)
	}

	_, err = divide(10, 0) // 使用空白标识符 `_` 忽略不关心的返回值
	if err != nil {
		fmt.Println("捕获到错误:", err)
	}

	// --- 调用结构体方法 ---
	fmt.Println("\n--- 结构体与方法 ---")
	r := Rect{Width: 10, Height: 5}
	fmt.Printf("矩形 r 的信息: %+v\n", r) // %+v 会打印字段名和值
	fmt.Println("矩形 r 的面积是:", r.Area()) //像调用对象方法一样调用
}
