package com.lightstore.user.interfaces.rest.model;

import lombok.Data;

import javax.validation.constraints.NotNull;

@Data
public class UserRegistrationRequest {
    @NotNull
    private String Username;
}