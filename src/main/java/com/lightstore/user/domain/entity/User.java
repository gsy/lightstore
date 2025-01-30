package com.lightstore.user.domain.entity;

import com.lightstore.user.domain.command.RegisterUser;
import com.lightstore.user.domain.command.handler.UserRegistrationHandler;
import com.lightstore.user.domain.shared.AggregateRoot;
import com.lightstore.user.domain.vo.UserID;
import org.springframework.context.ApplicationContext;


public class User extends AggregateRoot<User, UserID> {
    private String username;

    public User(ApplicationContext applicationContext) {
        super(applicationContext, new UserID());
    }

    public User(ApplicationContext applicationContext, UserID userID) {
        super(applicationContext, userID);
    }

    @Override
    public boolean sameIdentityAs(User other) {
        return other != null && entityID.sameValueAs(other.entityID);
    }

    @Override
    public UserID id() {
        return entityID;
    }

    @Override
    protected AggregateRootBehavior initialBehavior() {
        AggregateRootBehaviorBuilder builder = new AggregateRootBehaviorBuilder();
        builder.setCommandHandler(RegisterUser.class, getHandler(UserRegistrationHandler.class));
        return builder.build();
    }
}
