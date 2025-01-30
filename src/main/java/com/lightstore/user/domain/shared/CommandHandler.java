package com.lightstore.user.domain.shared;

import io.vavr.control.Either;

public interface CommandHandler<C extends Command, E extends Event, ID> {
    Either<CommandFailure, E> handle(C command, ID entityID);
}
