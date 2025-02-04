package com.lightstore.user.domain.command.handler;

import com.lightstore.user.domain.command.RegisterUser;
import com.lightstore.user.domain.entity.User;
import com.lightstore.user.domain.entity.UserRepository;
import com.lightstore.user.domain.event.UserRegistered;
import com.lightstore.user.domain.shared.CommandFailure;
import com.lightstore.user.domain.shared.CommandHandler;
import com.lightstore.user.domain.vo.UserID;
import io.vavr.control.Either;
import org.springframework.context.ApplicationContext;
import org.springframework.stereotype.Component;

@Component
public class UserRegistrationHandler implements CommandHandler<RegisterUser, UserRegistered, UserID> {

    private final ApplicationContext applicationContext;
    private final UserRepository repository;

    public UserRegistrationHandler(ApplicationContext applicationContext, UserRepository repository) {
        this.applicationContext = applicationContext;
        this.repository = repository;
    }

    @Override
    public Either<CommandFailure, UserRegistered> handle(RegisterUser command, UserID entityID) {
        User user = new User(applicationContext, entityID);
        repository.save(user);
        UserRegistered event = UserRegistered.eventOf(
                entityID,
                command.getUsername()
        );
        return Either.right(event);
    }
}
