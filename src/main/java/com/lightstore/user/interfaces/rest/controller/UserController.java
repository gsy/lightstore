package com.lightstore.user.interfaces.rest.controller;

import com.lightstore.user.application.UserApplication;
import com.lightstore.user.domain.command.RegisterUser;
import com.lightstore.user.domain.event.UserRegistered;
import com.lightstore.user.domain.shared.CommandFailure;
import com.lightstore.user.interfaces.rest.model.UserRegistrationRequest;
import com.lightstore.user.interfaces.rest.model.UserRegistrationResponse;
import io.vavr.control.Either;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;

import javax.validation.Valid;

@Controller
@RequestMapping(path="/v1/users")
public class UserController {
    private final UserApplication userApplication;

    public UserController(UserApplication application) {
        this.userApplication = application;
    }

    @PostMapping()
    public ResponseEntity<?> process(@Valid @RequestBody UserRegistrationRequest request) {

        RegisterUser userCommand = RegisterUser.commandOf(
                request.getUsername()
        );
        Either<CommandFailure, UserRegistered> result = this.userApplication.registerUser(userCommand);
        return result.fold(
                commandFailure -> ResponseEntity.badRequest().body(commandFailure.codes),
                event -> ResponseEntity.accepted().body(new UserRegistrationResponse(event.getUserID()))
        );
    }
}
