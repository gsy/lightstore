package com.lightstore.user.interfaces.rest.model;

import com.lightstore.user.domain.vo.UserID;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.Setter;

@AllArgsConstructor
@Getter
@Setter
public class UserRegistrationResponse {
    private UserID userID;
}
