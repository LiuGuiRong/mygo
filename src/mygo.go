package main

import (
    _"github.com/go-sql-driver/mysql"
    "database/sql"
    "fmt"
    "runtime"
    "os"
    "strconv" //convert number to string
    "github.com/widuu/goini" //config
    //"container/list"
    "strings"
    //"time"
    "labix.org/v2/mgo"
    //"labix.org/v2/mgo/bson"
    //"time"
    //"unicode"
    "reflect"
    "github.com/arnehormann/sqlinternals/mysqlinternals"
)

/// define 
const db_path = "./conf/db.ini"
const mysql_tag = "mysql"
const mongo_tag = "mongo"
const mysql_pattern_Alert = true

type DB_Cfg_Info struct {
    ip string
    port string
    passport string
    passwd string
    db_name string
}

func get_db_info(path string, name string) (info* DB_Cfg_Info) {
    conf := goini.SetConfig(path)

    username := conf.GetValue(name, "username")
    password := conf.GetValue(name, "password")
    hostname := conf.GetValue(name, "hostname")
    port := conf.GetValue(name, "port")
    db := conf.GetValue(name, "db")

    info = &DB_Cfg_Info{ip : hostname, port : port, passport : username, passwd : password, db_name : db}

    return
    /*
    return conf.GetValue(name, "username"),
        conf.GetValue(name, "password"),
        conf.GetValue(name, "hostname"),
        conf.GetValue(name, "port"),
        conf.GetValue(name, "db")
    */
}

func init_conf() (mysql_info, mongo_info * DB_Cfg_Info) {
    mysql_info = get_db_info(db_path, mysql_tag)
    mongo_info = get_db_info(db_path, mongo_tag)

    return
}

func init_tables() (names []string) {
    // closure pre-declare
    var waitting_input = func() {}
    var interact = func() {}

    waitting_input = func() () {
        var table_names string
        fmt.Println("Please enter the name of the table you want to import? (Separated by a comma or space)")
        _, err := fmt.Scanln(&table_names)
        checkErr(err)

        if table_names == "" {
            waitting_input()
        }

        names = strings.Split(table_names, ",")

        interact()

        return
    }

    interact = func() {
        var determine string

        fmt.Println("You input table names: ")
        for i := 0; i < len(names); i++ {
            if names[i] != "" {
                fmt.Println(i+1, names[i])
            }
        }

        fmt.Println("Are you sure import? (y or n)")
        _, err := fmt.Scanln(&determine)
        checkErr(err)
        switch {
        case determine == "y":
            fmt.Println("3Q, have a good day, sir~")
        case determine == "n":
            waitting_input()
        default:
            fmt.Println("unexpected input, please input again...")
            interact()
        }
    }

    waitting_input()

    return
}

func init_all() (mysql_info, mongo_info * DB_Cfg_Info, names []string) {

    mysql_info, mongo_info = init_conf()
    names =  init_tables()

    if mysql_info == nil {
        fmt.Println("3")
    }

    return
}

func read_from_mysql(mysql_info * DB_Cfg_Info, names []string) (rdata *map[string] *[][]interface{}, rcolume_name * map[string] *[]string, rcolume_type * map[string] *[]* reflect.Type) {
    url := mysql_info.passport + ":" + mysql_info.passwd + "@tcp(" + mysql_info.ip + ":" + mysql_info.port + ")/" + mysql_info.db_name + "?charset=utf8"

    fmt.Println(url)

    db, err := sql.Open(mysql_tag, url)
    checkErr(err)

    err = db.Ping()
    checkErr(err)

    fmt.Println("connect mysql successful")

    data := make(map[string] *[][]interface{})
    colume_name := make(map[string] *[]string)
    colume_type := make(map[string] *[]* reflect.Type)

    rdata = &data
    rcolume_name = &colume_name
    rcolume_type = &colume_type

    for i := 0; i < len(names); i++ {
        table_name := names[i]

        rows, err := db.Query("SELECT * FROM " + table_name)

        fmt.Println("load from mysql ...", table_name)

        checkErr(err)

        // Get column names
        columns, err := rows.Columns()
        if err != nil {
            panic(err.Error()) // proper error handling instead of panic in your app
        }

        colume_name[table_name] = &columns

        // get columns size
        columns_len := len(columns)
        rows_size := 0
        for rows.Next() {
            rows_size++
        }

        // get columns type
        cols_type_array := make([]* reflect.Type, len(columns))
        cols_type, err := mysqlinternals.Columns(rows)
        for i := range cols_type {
            refType, err := cols_type[i].ReflectGoType()
            checkErr(err)
            cols_type_array[i] = &refType
        }

        colume_type[table_name] = &cols_type_array


        sub_data := make([][]interface{}, rows_size)
        data[table_name] = &sub_data

        rows, err = db.Query("SELECT * FROM " + table_name)
        checkErr(err)

        // Make a slice for the values
        values := make([]sql.RawBytes, columns_len)
        // rows.Scan wants '[]interface{}' as an argument, so we must copy the
        // references into such a slice
        // See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
        scanArgs := make([]interface{}, len(values))

        for i := range values {
            scanArgs[i] = &values[i]
        }

        table_index := 0
        // Fetch rows
        for rows.Next() {
            // get RawBytes from data
            err = rows.Scan(scanArgs...)
            if err != nil {
                panic(err.Error()) // proper error handling instead of panic in your app
            }

            // Now do something with the data.
            // Here we just print each column as a string.
            sub_data[table_index] = make([]interface{}, columns_len + 1)

            var value string
            for i, col := range values{
                if col == nil {
                    value = "NULL"
                } else {
                    value = string(col)
                }
                sub_data[table_index][i] = value
            }

            table_index++
        }
    }
    fmt.Println(data)

    return
}

func checkErr(err error) {
    if err != nil {
        _, filename, lineno, ok := runtime.Caller(1)
        if ok {
            fmt.Fprintf(os.Stderr, "%v:%v: %v\n", filename, lineno, err)
        }
        panic(err)
    }
}

func write_to_mongo(mongo_info* DB_Cfg_Info, data *map[string] *[][]interface{}, colume_name *map[string] *[]string, colume_type * map[string] *[]* reflect.Type) {
    //url := "mongodb://" + mongo_info.passport + ":" + mongo_info.passwd + "@" + mongo_info.ip + ":" + mongo_info.port
    url := mongo_info.ip + ":27017"
    session, err := mgo.Dial(url)
    checkErr(err)
    defer session.Close()

    session.SetMode(mgo.Monotonic, true)

    db := session.DB(mongo_info.db_name)

    check_null := func(value *string, default_value string, colletion_name * string, row_name * string, columns_index int) (b bool) {
        if *value == "NULL" {
            *value = default_value

            if mysql_pattern_Alert == true {
                fmt.Println("warn! data from mysql is nil , table name:", *colletion_name, " row name:", *row_name, " columns index:", columns_index)
            }

            b = true
        }

        b = false

        return
    }

    for colletion_name, collection := range *data {
        c := db.C(colletion_name)
        if c != nil {
            err = c.DropCollection()
            //checkErr(err)
        }

        // get name
        colume_name_raw := *colume_name
        colume_index := colume_name_raw[colletion_name]

        colume_type_raw := *colume_type
        colume_type := *colume_type_raw[colletion_name]

        for columns_index, row_info := range *collection {

            formate_info := make(map[string] interface{}, len(*colume_index))

            for ib, name := range *colume_index{
                v := row_info[ib].(string)
                t := (*colume_type[ib]).Kind()

                switch {
                default:
                    fmt.Printf("Unexpected type %T\n", t)
                case reflect.Int == t ||
                    reflect.Int8 == t ||
                    reflect.Int16 == t ||
                    reflect.Int32 == t ||
                    reflect.Uint == t ||
                    reflect.Uint8 == t ||
                    reflect.Uint16 == t ||
                    reflect.Uint32 == t :

                    check_null(&v, "0", &colletion_name, &name, columns_index)

                    i, err := strconv.ParseInt(v, 10, 32)
                    checkErr(err)
                    formate_info[name] = i
                case reflect.Int64 == t ||
                    reflect.Uint64 == t:

                    check_null(&v, "0", &colletion_name, &name, columns_index)

                    i, err := strconv.ParseInt(v, 10, 64)
                    checkErr(err)
                    formate_info[name] = i
                case reflect.Float32 == t:

                    check_null(&v, "0.0", &colletion_name, &name, columns_index)

                    i, err := strconv.ParseFloat(v, 32)
                    checkErr(err)
                    formate_info[name] = i
                case reflect.Float64 == t:

                    check_null(&v, "0.0", &colletion_name, &name, columns_index)

                    i, err := strconv.ParseFloat(v, 64)
                    checkErr(err)
                    formate_info[name] = i
                case reflect.Bool == t:

                    check_null(&v, "true", &colletion_name, &name, columns_index)

                    i, err := strconv.ParseBool(v)
                    checkErr(err)
                    formate_info[name] = i
                case reflect.String == t:

                    check_null(&v, "NULL", &colletion_name, &name, columns_index)

                    formate_info[name] = v
                }
            }

            err = c.Insert(&formate_info)
            checkErr(err)
        }

        index := mgo.Index{
            Key:        []string{"id","player"},
            Unique:     false,
            DropDups:   false,
            Background: false,
            Sparse:     false,
        }
        err = c.EnsureIndex(index)

    }

    db.Close()
}

func import_() {
    if mysql_info, mongo_info, names := init_all(); len(names) > 0 {
        data, colume_name, colume_type := read_from_mysql(mysql_info, names)
        write_to_mongo(mongo_info, data, colume_name, colume_type)
    }
}

func main() {
    import_()
}
