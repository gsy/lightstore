package com.lightstore.user.application;

import com.lightstore.user.domain.command.RegisterUser;
import com.lightstore.user.domain.event.UserRegistered;
import com.lightstore.user.domain.shared.CommandFailure;
import io.vavr.control.Either;

public interface UserApplication {
    Either<CommandFailure, UserRegistered> registerUser(RegisterUser command);
}
