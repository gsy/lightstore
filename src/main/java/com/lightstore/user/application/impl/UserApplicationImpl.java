package com.lightstore.user.application.impl;

import com.lightstore.user.application.UserApplication;
import com.lightstore.user.domain.command.RegisterUser;
import com.lightstore.user.domain.entity.User;
import com.lightstore.user.domain.event.UserRegistered;
import com.lightstore.user.domain.shared.CommandFailure;
import io.vavr.control.Either;
import org.springframework.context.ApplicationContext;
import org.springframework.stereotype.Service;

@Service
public class UserApplicationImpl implements UserApplication {

    private final ApplicationContext applicationContext;

    UserApplicationImpl(ApplicationContext applicationContext) {
        this.applicationContext = applicationContext;
    }

    @Override
    public Either<CommandFailure, UserRegistered> registerUser(RegisterUser registerUserCommand) {
        User user = new User(this.applicationContext);
        return user.handle(registerUserCommand);
    }
}