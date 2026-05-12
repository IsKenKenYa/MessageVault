package imken.messagevault.mobile.data.backup

import android.content.ContentProviderOperation
import android.content.ContentResolver
import android.content.Context
import android.provider.ContactsContract
import imken.messagevault.sdk.backup.model.ContactData
import imken.messagevault.sdk.backup.writer.ContactWriter
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import timber.log.Timber

class AndroidContactWriter(
    private val context: Context
) : ContactWriter {

    private val contentResolver: ContentResolver = context.contentResolver

    override suspend fun write(contacts: List<ContactData>): Int = withContext(Dispatchers.IO) {
        var restoredCount = 0
        val totalCount = contacts.size
        Timber.i("[Mobile] INFO [Restore] Starting contact restore, total: $totalCount")

        if (!hasContactsPermissions()) {
            Timber.e("[Mobile] ERROR [Restore] Missing contacts permissions")
            return@withContext 0
        }

        contacts.forEachIndexed { index, contact ->
            try {
                val operations = ArrayList<ContentProviderOperation>()

                val rawContactInsertIndex = operations.size
                operations.add(ContentProviderOperation.newInsert(ContactsContract.RawContacts.CONTENT_URI)
                    .withValue(ContactsContract.RawContacts.ACCOUNT_TYPE, null)
                    .withValue(ContactsContract.RawContacts.ACCOUNT_NAME, null)
                    .build())

                operations.add(ContentProviderOperation.newInsert(ContactsContract.Data.CONTENT_URI)
                    .withValueBackReference(ContactsContract.Data.RAW_CONTACT_ID, rawContactInsertIndex)
                    .withValue(ContactsContract.Data.MIMETYPE, ContactsContract.CommonDataKinds.StructuredName.CONTENT_ITEM_TYPE)
                    .withValue(ContactsContract.CommonDataKinds.StructuredName.DISPLAY_NAME, contact.name)
                    .build())

                contact.phoneNumbers.forEach { phoneNumber ->
                    operations.add(ContentProviderOperation.newInsert(ContactsContract.Data.CONTENT_URI)
                        .withValueBackReference(ContactsContract.Data.RAW_CONTACT_ID, rawContactInsertIndex)
                        .withValue(ContactsContract.Data.MIMETYPE, ContactsContract.CommonDataKinds.Phone.CONTENT_ITEM_TYPE)
                        .withValue(ContactsContract.CommonDataKinds.Phone.NUMBER, phoneNumber)
                        .withValue(ContactsContract.CommonDataKinds.Phone.TYPE, ContactsContract.CommonDataKinds.Phone.TYPE_MOBILE)
                        .build())
                }

                try {
                    val results = contentResolver.applyBatch(ContactsContract.AUTHORITY, operations)
                    if (results.isNotEmpty()) {
                        restoredCount++
                    }
                } catch (e: Exception) {
                    Timber.e(e, "[Mobile] ERROR [Restore] Failed to restore contact: batch operation exception: ${e.message}")
                }
            } catch (e: Exception) {
                Timber.e(e, "[Mobile] ERROR [Restore] Failed to restore contact: ID=${contact.id}, name=${contact.name}, ${e.message}")
            }
        }

        Timber.i("[Mobile] INFO [Restore] Contact restore complete: success=$restoredCount, total=$totalCount")
        restoredCount
    }

    private fun hasContactsPermissions(): Boolean {
        val readPermission = android.Manifest.permission.READ_CONTACTS
        val writePermission = android.Manifest.permission.WRITE_CONTACTS

        val readGranted = context.checkSelfPermission(readPermission) == android.content.pm.PackageManager.PERMISSION_GRANTED
        val writeGranted = context.checkSelfPermission(writePermission) == android.content.pm.PackageManager.PERMISSION_GRANTED

        if (!readGranted || !writeGranted) {
            Timber.e("[Mobile] ERROR [Restore] Missing contacts permissions: READ_CONTACTS=$readGranted, WRITE_CONTACTS=$writeGranted")
            return false
        }

        return true
    }
}
