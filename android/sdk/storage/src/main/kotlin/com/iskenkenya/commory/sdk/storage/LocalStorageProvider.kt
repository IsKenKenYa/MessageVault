package com.iskenkenya.commory.sdk.storage

import android.content.Context
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.withContext
import java.io.File
import java.net.URLConnection

class LocalStorageProvider(
    private val context: Context
) : StorageProvider {

    private val _isAvailable = MutableStateFlow(false)
    override val isAvailable: StateFlow<Boolean> = _isAvailable

    private val baseDir: File = context.getExternalFilesDir(null)
        ?: context.filesDir

    init {
        refreshAvailability()
    }

    private fun refreshAvailability() {
        _isAvailable.value = baseDir.exists() && baseDir.canWrite()
    }

    override suspend fun save(data: ByteArray, path: String): StorageResult =
        withContext(Dispatchers.IO) {
            try {
                val file = File(baseDir, path)
                file.parentFile?.mkdirs()
                file.writeBytes(data)
                refreshAvailability()
                StorageResult.Success(path = file.absolutePath)
            } catch (e: Exception) {
                StorageResult.Error(code = "SAVE_FAILED", message = e.message ?: "Unknown error")
            }
        }

    override suspend fun load(path: String): StorageResult =
        withContext(Dispatchers.IO) {
            try {
                val file = File(baseDir, path)
                if (!file.exists()) {
                    return@withContext StorageResult.Error(code = "FILE_NOT_FOUND", message = "File not found: $path")
                }
                val data = file.readBytes()
                StorageResult.Success(path = file.absolutePath, data = data)
            } catch (e: Exception) {
                StorageResult.Error(code = "LOAD_FAILED", message = e.message ?: "Unknown error")
            }
        }

    override suspend fun list(prefix: String): List<StorageFileInfo> =
        withContext(Dispatchers.IO) {
            val searchDir = if (prefix.isNotEmpty()) {
                File(baseDir, prefix)
            } else {
                baseDir
            }
            if (!searchDir.exists() || !searchDir.isDirectory) {
                return@withContext emptyList()
            }
            searchDir.walkTopDown()
                .filter { it.isFile }
                .map { file ->
                    val relativePath = file.relativeTo(baseDir).path
                    StorageFileInfo(
                        path = relativePath,
                        size = file.length(),
                        lastModified = file.lastModified(),
                        mimeType = guessMimeType(file)
                    )
                }
                .toList()
        }

    override suspend fun delete(path: String): StorageResult =
        withContext(Dispatchers.IO) {
            try {
                val file = File(baseDir, path)
                if (!file.exists()) {
                    return@withContext StorageResult.Error(code = "FILE_NOT_FOUND", message = "File not found: $path")
                }
                val deleted = file.delete()
                if (deleted) {
                    refreshAvailability()
                    StorageResult.Success(path = path)
                } else {
                    StorageResult.Error(code = "DELETE_FAILED", message = "Failed to delete: $path")
                }
            } catch (e: Exception) {
                StorageResult.Error(code = "DELETE_FAILED", message = e.message ?: "Unknown error")
            }
        }

    override suspend fun getStorageInfo(): StorageInfo =
        withContext(Dispatchers.IO) {
            refreshAvailability()
            val totalSpace = baseDir.totalSpace
            val availableSpace = baseDir.usableSpace
            val usedSpace = totalSpace - availableSpace
            StorageInfo(
                totalSpace = totalSpace,
                usedSpace = usedSpace,
                availableSpace = availableSpace
            )
        }

    private fun guessMimeType(file: File): String? {
        val name = file.name
        return try {
            URLConnection.guessContentTypeFromName(name)
        } catch (e: Exception) {
            null
        }
    }
}
