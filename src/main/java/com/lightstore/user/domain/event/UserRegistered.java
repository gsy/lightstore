package com.lightstore.user.domain.event;


import com.lightstore.user.domain.vo.UserID;
import lombok.Getter;
import lombok.Value;


@Value(staticConstructor = "eventOf")
public class UserRegistered implements UserEvent {
    private final UserID userID;
    @Getter
    private final String username;

    @Override
    public UserID getUserID() {
        return userID;
    }
}
