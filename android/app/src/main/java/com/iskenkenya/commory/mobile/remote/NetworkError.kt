package com.iskenkenya.commory.mobile.remote

import java.io.IOException
import java.net.SocketTimeoutException
import java.net.UnknownHostException
import javax.net.ssl.SSLException

sealed class NetworkError(message: String, cause: Throwable? = null) : IOException(message, cause) {
    class Offline(cause: Throwable? = null) : NetworkError("network unavailable", cause)
    class Timeout(cause: Throwable? = null) : NetworkError("network timeout", cause)
    class Ssl(cause: Throwable? = null) : NetworkError("ssl error", cause)
    class Unauthorized(message: String = "unauthorized") : NetworkError(message)
    class Server(val statusCode: Int, message: String) : NetworkError(message)
    class Validation(message: String) : NetworkError(message)
    class Unknown(cause: Throwable) : NetworkError(cause.message ?: "network error", cause)
}

class NetworkException(val error: NetworkError) : IOException(error.message, error)

fun Throwable.asNetworkException(): NetworkException {
    return when (this) {
        is NetworkException -> this
        is UnknownHostException -> NetworkException(NetworkError.Offline(this))
        is SocketTimeoutException -> NetworkException(NetworkError.Timeout(this))
        is SSLException -> NetworkException(NetworkError.Ssl(this))
        is IOException -> NetworkException(NetworkError.Unknown(this))
        else -> NetworkException(NetworkError.Unknown(this))
    }
}
