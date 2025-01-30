package com.lightstore.user.domain.command.handler;

import com.lightstore.user.domain.command.RegisterUser;
import com.lightstore.user.domain.event.UserEvent;
import com.lightstore.user.domain.event.UserRegistered;
import com.lightstore.user.domain.shared.Command;
import com.lightstore.user.domain.shared.CommandFailure;
import com.lightstore.user.domain.shared.CommandHandler;
import com.lightstore.user.domain.vo.UserID;
import io.vavr.control.Either;
import org.springframework.stereotype.Component;

@Component
public class UserRegistrationHandler implements CommandHandler<RegisterUser, UserRegistered, UserID> {

    @Override
    public Either<CommandFailure, UserRegistered> handle(RegisterUser command, UserID entityID) {
        UserRegistered event = UserRegistered.eventOf(
                entityID,
                command.getUsername()
        );
        return Either.right(event);
    }
}
