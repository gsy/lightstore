package com.lightstore.user.domain.shared;

import io.vavr.control.Either;
import org.springframework.context.ApplicationContext;

import java.io.Serializable;
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

public abstract class AggregateRoot<E, ID extends Serializable> implements Entity<E, ID>{
    public final ID entityID;
    private final ApplicationContext applicationContext;
    private AggregateRootBehavior behavior;

    protected AggregateRoot(ApplicationContext applicationContext, ID entityID) {
        this.applicationContext = applicationContext;
        this.entityID = entityID;
        this.behavior = initialBehavior();
    }

    public <A extends Command, B extends Event> Either<CommandFailure, B> handle(A command) {
        CommandHandler<A, B, ID> commandHandler = (CommandHandler<A, B, ID>) behavior.handlers.get(command.getClass());
        return commandHandler.handle(command, entityID);
    }

    protected <A extends Command, B extends Event> CommandHandler<A, B, ID> getHandler(Class<? extends CommandHandler> commandHandlerClass) {
        return applicationContext.getBean(commandHandlerClass);
    }

    protected abstract AggregateRootBehavior initialBehavior();

    public class AggregateRootBehavior<ID> {
        protected final Map<Class<? extends Command>, CommandHandler<? extends Command, ? extends  Event, ID>> handlers;

        public AggregateRootBehavior(Map<Class<? extends Command>, CommandHandler<? extends Command, ? extends  Event, ID>> handlers) {
            this.handlers = Collections.unmodifiableMap(handlers);
        }
    }

    public class AggregateRootBehaviorBuilder<ID> {
        private final Map<Class<? extends Command>, CommandHandler<? extends Command, ? extends Event, ID>> handlers = new HashMap<>();

        public <A extends Command, B extends Event> AggregateRootBehaviorBuilder setCommandHandler(Class<A> commandClass, CommandHandler<A, B, ID> handler) {
            handlers.put(commandClass, handler);
            return this;
        }

        public AggregateRootBehavior build() {
            return new AggregateRootBehavior(handlers);
        }
    }
}
