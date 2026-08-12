fnc bagi(a, b) {
    if b == 0 {
        error "tidak bisa bagi nol"
    }
    return a / b
}

fnc main {
    try hasil = bagi(10, 0)
    if hasil == _ {
        print("gagal bagi, error ketangkep!")
    }

    normal = bagi(10, 2)
    print(normal)
}

main()
