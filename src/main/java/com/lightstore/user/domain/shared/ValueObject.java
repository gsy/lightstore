package com.lightstore.user.domain.shared;

import java.io.Serializable;

public interface ValueObject<T> extends Serializable {
    boolean sameValueAs(T other);
}
