package com.lightstore.user;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.ComponentScan;

@SpringBootApplication
@ComponentScan
public class Applicaiton {
    public static void main(String[] args) {
        SpringApplication.run(Applicaiton.class, args);
    }
}
