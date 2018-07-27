package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

const DB_HOST = "127.0.0.1"
const DB_PORT = "3306"
const DB_NAME = "golang"
const DB_USER = "phpmyadmin"
const DB_PASS = "123456"

func CreateDatabse(w http.ResponseWriter, r *http.Request) {

	// Connect
	db, err := sql.Open("mysql", DB_USER+":"+DB_PASS+"@tcp("+DB_HOST+":"+DB_PORT+")/"+DB_NAME)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	} else {git commit -m "First Go Project"
		fmt.Println("Connect")
	}

	defer db.Close()

	// Create User Table
	_, err = db.Exec("CREATE TABLE `users` (`id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL, `email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL, `password` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL, `created_at` timestamp NULL DEFAULT NULL) CHARSET=utf8mb4;")
	if err != nil {
		fmt.Fprintf(w, "NOT FOUND")
		return
	} else {
		fmt.Println("Successful migrate users table")

		// Alter Table

		_, err = db.Exec("ALTER TABLE `users` ADD PRIMARY KEY (`id`), ADD UNIQUE KEY `users_email_unique` (`email`);")
		if err != nil {
			fmt.Println(err)
		}

	}

	// Create access_tokens Table

	_, err = db.Exec("CREATE TABLE `access_tokens` (`id` integer(11) COLLATE utf8mb4_unicode_ci NOT NULL,`user_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,`access_token` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL, `created_at` timestamp NULL DEFAULT NULL) CHARSET=utf8mb4;")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Successful migrate access_tokens table")

		_, err = db.Exec("ALTER TABLE `access_tokens` ADD PRIMARY KEY (`id`)")
		if err != nil {
			fmt.Println(err)
		}

		_, err = db.Exec("ALTER TABLE `access_tokens` MODIFY COLUMN id INT auto_increment")
		if err != nil {
			fmt.Println(err)
		}

	}

	// Create records Table
	_, err = db.Exec("CREATE TABLE `records` (`id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,`user_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,`fork_id` char(36) COLLATE utf8mb4_unicode_ci NULL,`freeze` boolean default 0 NOT NULL, `created_at` timestamp NULL DEFAULT NULL, FOREIGN KEY (user_id) REFERENCES users(id)  ON DELETE CASCADE ) CHARSET=utf8mb4;")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Successful migrate records table")

		_, err = db.Exec("ALTER TABLE `records` ADD PRIMARY KEY (`id`)")
		if err != nil {
			fmt.Println(err)
		}

	}

	// Create recordd_history Table
	_, err = db.Exec("CREATE TABLE `records_history` (`id` integer(11) COLLATE utf8mb4_unicode_ci NOT NULL,`record_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL,`step` integer(2) COLLATE utf8mb4_unicode_ci NOT NULL,`operation` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,`op_value` integer(10) COLLATE utf8mb4_unicode_ci NULL, `value` integer(10) COLLATE utf8mb4_unicode_ci NOT NULL, `created_at` timestamp NULL DEFAULT NULL,  FOREIGN KEY (record_id) REFERENCES records(id) ON DELETE CASCADE ) CHARSET=utf8mb4;")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Successful migrate records table")

		_, err = db.Exec("ALTER TABLE `records_history` ADD PRIMARY KEY (`id`)")
		if err != nil {
			fmt.Println(err)
		}
		_, err = db.Exec("ALTER TABLE `records_history` MODIFY COLUMN id INT auto_increment")
		if err != nil {
			fmt.Println(err)
		}
	}

}
