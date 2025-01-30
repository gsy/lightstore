package com.lightstore.user.domain.command;

import lombok.Value;

@Value(staticConstructor = "commandOf")
public class RegisterUser implements UserCommand{
    private final String username;
}