orang {
    nama
    umur
}

fnc main {
    orang {
        nama = "Budi"
        umur = 20
    }
    budi = orang
    print(budi.nama)
    print(budi.umur)

    angka = [1, 2, 3]
    print(angka[0])
    print(len(angka))

    data = {"kota": "Jakarta"}
    print(data["kota"])
}

main()
