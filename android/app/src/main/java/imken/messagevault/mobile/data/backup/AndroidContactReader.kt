package imken.messagevault.mobile.data.backup

import android.content.Context
import android.os.Build
import android.provider.ContactsContract
import imken.messagevault.sdk.backup.model.Contact
import imken.messagevault.sdk.backup.reader.ContactReader
import timber.log.Timber

class AndroidContactReader(private val context: Context) : ContactReader {

    override suspend fun readContacts(): List<Contact>? {
        val contacts = mutableListOf<Contact>()

        try {
            val permissionStatus = context.checkSelfPermission(android.Manifest.permission.READ_CONTACTS)
            if (permissionStatus != android.content.pm.PackageManager.PERMISSION_GRANTED) {
                Timber.e("[Mobile] ERROR [Backup] 备份联系人失败: 没有 READ_CONTACTS 权限")
                return null
            }

            Timber.d("[Mobile] DEBUG [Backup] 开始读取联系人...")

            val uri = ContactsContract.Contacts.CONTENT_URI
            val projection = arrayOf(
                ContactsContract.Contacts._ID,
                ContactsContract.Contacts.DISPLAY_NAME,
                ContactsContract.Contacts.HAS_PHONE_NUMBER,
                ContactsContract.Contacts.PHOTO_URI
            )
            val sortOrder = "${ContactsContract.Contacts.DISPLAY_NAME} ASC"

            context.contentResolver.query(uri, projection, null, null, sortOrder)?.use { cursor ->
                Timber.d("[Mobile] DEBUG [Backup] 找到 ${cursor.count} 个联系人")

                val idColumn = cursor.getColumnIndex(ContactsContract.Contacts._ID)
                val nameColumn = cursor.getColumnIndex(ContactsContract.Contacts.DISPLAY_NAME)
                val hasPhoneColumn = cursor.getColumnIndex(ContactsContract.Contacts.HAS_PHONE_NUMBER)
                val photoUriColumn = cursor.getColumnIndex(ContactsContract.Contacts.PHOTO_URI)

                val batchSize = 50
                var contactsProcessed = 0

                while (cursor.moveToNext()) {
                    try {
                        val id = if (idColumn != -1) cursor.getLong(idColumn) else 0
                        val name = if (nameColumn != -1) cursor.getString(nameColumn) else ""
                        val hasPhone = if (hasPhoneColumn != -1) cursor.getInt(hasPhoneColumn) > 0 else false

                        val phoneNumbers = mutableListOf<String>()
                        val emails = mutableListOf<String>()
                        val addresses = mutableListOf<Contact.Address>()
                        val websites = mutableListOf<String>()
                        val socialProfiles = mutableListOf<Contact.SocialProfile>()
                        val relationships = mutableListOf<Contact.Relationship>()
                        var note: String? = null
                        val groups = mutableListOf<String>()
                        val events = mutableListOf<Contact.Event>()

                        if (hasPhone) {
                            val phoneUri = ContactsContract.CommonDataKinds.Phone.CONTENT_URI
                            val phoneProjection = arrayOf(
                                ContactsContract.CommonDataKinds.Phone.NUMBER,
                                ContactsContract.CommonDataKinds.Phone.TYPE
                            )
                            val phoneSelection = "${ContactsContract.CommonDataKinds.Phone.CONTACT_ID} = ?"
                            val phoneSelectionArgs = arrayOf(id.toString())

                            context.contentResolver.query(phoneUri, phoneProjection, phoneSelection, phoneSelectionArgs, null)?.use { phoneCursor ->
                                val phoneNumberColumn = phoneCursor.getColumnIndex(ContactsContract.CommonDataKinds.Phone.NUMBER)

                                while (phoneCursor.moveToNext()) {
                                    val phoneNumber = if (phoneNumberColumn != -1) phoneCursor.getString(phoneNumberColumn) else ""
                                    if (phoneNumber.isNotBlank()) {
                                        phoneNumbers.add(phoneNumber)
                                    }
                                }
                            }
                        }

                        val emailUri = ContactsContract.CommonDataKinds.Email.CONTENT_URI
                        val emailProjection = arrayOf(
                            ContactsContract.CommonDataKinds.Email.ADDRESS,
                            ContactsContract.CommonDataKinds.Email.TYPE
                        )
                        val emailSelection = "${ContactsContract.CommonDataKinds.Email.CONTACT_ID} = ?"
                        val emailSelectionArgs = arrayOf(id.toString())

                        context.contentResolver.query(emailUri, emailProjection, emailSelection, emailSelectionArgs, null)?.use { emailCursor ->
                            val emailAddressColumn = emailCursor.getColumnIndex(ContactsContract.CommonDataKinds.Email.ADDRESS)

                            while (emailCursor.moveToNext()) {
                                val emailAddress = if (emailAddressColumn != -1) emailCursor.getString(emailAddressColumn) else ""
                                if (emailAddress.isNotBlank()) {
                                    emails.add(emailAddress)
                                }
                            }
                        }

                        val addressUri = ContactsContract.CommonDataKinds.StructuredPostal.CONTENT_URI
                        val addressProjection = arrayOf(
                            ContactsContract.CommonDataKinds.StructuredPostal.FORMATTED_ADDRESS,
                            ContactsContract.CommonDataKinds.StructuredPostal.TYPE
                        )
                        val addressSelection = "${ContactsContract.CommonDataKinds.StructuredPostal.CONTACT_ID} = ?"
                        val addressSelectionArgs = arrayOf(id.toString())

                        context.contentResolver.query(addressUri, addressProjection, addressSelection, addressSelectionArgs, null)?.use { addressCursor ->
                            val addressColumn = addressCursor.getColumnIndex(ContactsContract.CommonDataKinds.StructuredPostal.FORMATTED_ADDRESS)
                            val addressTypeColumn = addressCursor.getColumnIndex(ContactsContract.CommonDataKinds.StructuredPostal.TYPE)

                            while (addressCursor.moveToNext()) {
                                val address = if (addressColumn != -1) addressCursor.getString(addressColumn) else ""
                                val addressType = if (addressTypeColumn != -1) {
                                    when (addressCursor.getInt(addressTypeColumn)) {
                                        ContactsContract.CommonDataKinds.StructuredPostal.TYPE_HOME -> "家庭"
                                        ContactsContract.CommonDataKinds.StructuredPostal.TYPE_WORK -> "工作"
                                        else -> "其他"
                                    }
                                } else "其他"

                                if (address.isNotBlank()) {
                                    addresses.add(Contact.Address(addressType, address))
                                }
                            }
                        }

                        val noteUri = ContactsContract.Data.CONTENT_URI
                        val noteProjection = arrayOf(
                            ContactsContract.CommonDataKinds.Note.NOTE
                        )
                        val noteSelection = "${ContactsContract.Data.CONTACT_ID} = ? AND ${ContactsContract.Data.MIMETYPE} = ?"
                        val noteSelectionArgs = arrayOf(
                            id.toString(),
                            ContactsContract.CommonDataKinds.Note.CONTENT_ITEM_TYPE
                        )

                        context.contentResolver.query(noteUri, noteProjection, noteSelection, noteSelectionArgs, null)?.use { noteCursor ->
                            val noteColumn = noteCursor.getColumnIndex(ContactsContract.CommonDataKinds.Note.NOTE)

                            if (noteCursor.moveToFirst()) {
                                note = if (noteColumn != -1) noteCursor.getString(noteColumn) else null
                            }
                        }

                        val websiteUri = ContactsContract.Data.CONTENT_URI
                        val websiteProjection = arrayOf(
                            ContactsContract.CommonDataKinds.Website.URL
                        )
                        val websiteSelection = "${ContactsContract.Data.CONTACT_ID} = ? AND ${ContactsContract.Data.MIMETYPE} = ?"
                        val websiteSelectionArgs = arrayOf(
                            id.toString(),
                            ContactsContract.CommonDataKinds.Website.CONTENT_ITEM_TYPE
                        )

                        context.contentResolver.query(websiteUri, websiteProjection, websiteSelection, websiteSelectionArgs, null)?.use { websiteCursor ->
                            val urlColumn = websiteCursor.getColumnIndex(ContactsContract.CommonDataKinds.Website.URL)

                            while (websiteCursor.moveToNext()) {
                                val url = if (urlColumn != -1) websiteCursor.getString(urlColumn) else ""
                                if (url.isNotBlank()) {
                                    websites.add(url)
                                }
                            }
                        }

                        val eventUri = ContactsContract.Data.CONTENT_URI
                        val eventProjection = arrayOf(
                            ContactsContract.CommonDataKinds.Event.START_DATE,
                            ContactsContract.CommonDataKinds.Event.TYPE
                        )
                        val eventSelection = "${ContactsContract.Data.CONTACT_ID} = ? AND ${ContactsContract.Data.MIMETYPE} = ?"
                        val eventSelectionArgs = arrayOf(
                            id.toString(),
                            ContactsContract.CommonDataKinds.Event.CONTENT_ITEM_TYPE
                        )

                        context.contentResolver.query(eventUri, eventProjection, eventSelection, eventSelectionArgs, null)?.use { eventCursor ->
                            val dateColumn = eventCursor.getColumnIndex(ContactsContract.CommonDataKinds.Event.START_DATE)
                            val typeColumn = eventCursor.getColumnIndex(ContactsContract.CommonDataKinds.Event.TYPE)

                            while (eventCursor.moveToNext()) {
                                val date = if (dateColumn != -1) eventCursor.getString(dateColumn) else ""
                                val type = if (typeColumn != -1) {
                                    when (eventCursor.getInt(typeColumn)) {
                                        ContactsContract.CommonDataKinds.Event.TYPE_BIRTHDAY -> "生日"
                                        ContactsContract.CommonDataKinds.Event.TYPE_ANNIVERSARY -> "纪念日"
                                        else -> "其他"
                                    }
                                } else "其他"

                                if (date.isNotBlank()) {
                                    events.add(Contact.Event(type, date))
                                }
                            }
                        }

                        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
                            val groupUri = ContactsContract.Data.CONTENT_URI
                            val groupProjection = arrayOf(
                                ContactsContract.CommonDataKinds.GroupMembership.GROUP_ROW_ID
                            )
                            val groupSelection = "${ContactsContract.Data.CONTACT_ID} = ? AND ${ContactsContract.Data.MIMETYPE} = ?"
                            val groupSelectionArgs = arrayOf(
                                id.toString(),
                                ContactsContract.CommonDataKinds.GroupMembership.CONTENT_ITEM_TYPE
                            )

                            context.contentResolver.query(groupUri, groupProjection, groupSelection, groupSelectionArgs, null)?.use { groupCursor ->
                                val groupIdColumn = groupCursor.getColumnIndex(ContactsContract.CommonDataKinds.GroupMembership.GROUP_ROW_ID)

                                while (groupCursor.moveToNext()) {
                                    val groupId = if (groupIdColumn != -1) groupCursor.getLong(groupIdColumn) else -1

                                    if (groupId != -1L) {
                                        val groupNameUri = ContactsContract.Groups.CONTENT_URI
                                        val groupNameProjection = arrayOf(
                                            ContactsContract.Groups.TITLE
                                        )
                                        val groupNameSelection = "${ContactsContract.Groups._ID} = ?"
                                        val groupNameSelectionArgs = arrayOf(groupId.toString())

                                        context.contentResolver.query(groupNameUri, groupNameProjection, groupNameSelection, groupNameSelectionArgs, null)?.use { groupNameCursor ->
                                            val groupNameColumn = groupNameCursor.getColumnIndex(ContactsContract.Groups.TITLE)

                                            if (groupNameCursor.moveToFirst()) {
                                                val groupName = if (groupNameColumn != -1) groupNameCursor.getString(groupNameColumn) else ""
                                                if (groupName.isNotBlank()) {
                                                    groups.add(groupName)
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }

                        val relationUri = ContactsContract.Data.CONTENT_URI
                        val relationProjection = arrayOf(
                            ContactsContract.CommonDataKinds.Relation.NAME,
                            ContactsContract.CommonDataKinds.Relation.TYPE
                        )
                        val relationSelection = "${ContactsContract.Data.CONTACT_ID} = ? AND ${ContactsContract.Data.MIMETYPE} = ?"
                        val relationSelectionArgs = arrayOf(
                            id.toString(),
                            ContactsContract.CommonDataKinds.Relation.CONTENT_ITEM_TYPE
                        )

                        context.contentResolver.query(relationUri, relationProjection, relationSelection, relationSelectionArgs, null)?.use { relationCursor ->
                            val nameColumn = relationCursor.getColumnIndex(ContactsContract.CommonDataKinds.Relation.NAME)
                            val typeColumn = relationCursor.getColumnIndex(ContactsContract.CommonDataKinds.Relation.TYPE)

                            while (relationCursor.moveToNext()) {
                                val relName = if (nameColumn != -1) relationCursor.getString(nameColumn) else ""
                                val relType = if (typeColumn != -1) {
                                    when (relationCursor.getInt(typeColumn)) {
                                        ContactsContract.CommonDataKinds.Relation.TYPE_SPOUSE -> "配偶"
                                        ContactsContract.CommonDataKinds.Relation.TYPE_CHILD -> "子女"
                                        ContactsContract.CommonDataKinds.Relation.TYPE_PARENT -> "父母"
                                        else -> "其他"
                                    }
                                } else "其他"

                                if (relName.isNotBlank()) {
                                    relationships.add(Contact.Relationship(relType, relName))
                                }
                            }
                        }

                        val socialUri = ContactsContract.Data.CONTENT_URI
                        val socialProjection = arrayOf(
                            ContactsContract.Data.MIMETYPE,
                            ContactsContract.Data.DATA1
                        )
                        val socialSelection = "${ContactsContract.Data.CONTACT_ID} = ? AND ${ContactsContract.Data.MIMETYPE} IN (?, ?, ?)"
                        val socialSelectionArgs = arrayOf(
                            id.toString(),
                            "vnd.android.cursor.item/com.whatsapp.profile",
                            "vnd.android.cursor.item/com.facebook.profile",
                            "vnd.android.cursor.item/com.twitter.android.profile"
                        )

                        context.contentResolver.query(socialUri, socialProjection, socialSelection, socialSelectionArgs, null)?.use { socialCursor ->
                            val mimeTypeColumn = socialCursor.getColumnIndex(ContactsContract.Data.MIMETYPE)
                            val data1Column = socialCursor.getColumnIndex(ContactsContract.Data.DATA1)

                            while (socialCursor.moveToNext()) {
                                val mimeType = if (mimeTypeColumn != -1) socialCursor.getString(mimeTypeColumn) else ""
                                val data = if (data1Column != -1) socialCursor.getString(data1Column) else ""

                                if (data.isNotBlank()) {
                                    val socialType = when {
                                        mimeType.contains("whatsapp") -> "WhatsApp"
                                        mimeType.contains("facebook") -> "Facebook"
                                        mimeType.contains("twitter") -> "Twitter"
                                        else -> "其他"
                                    }

                                    socialProfiles.add(Contact.SocialProfile(socialType, data))
                                }
                            }
                        }

                        if (name.isNotBlank() || phoneNumbers.isNotEmpty() || emails.isNotEmpty()) {
                            val contact = Contact(
                                id = id,
                                name = name,
                                phoneNumbers = phoneNumbers,
                                emails = if (emails.isNotEmpty()) emails else null,
                                addresses = if (addresses.isNotEmpty()) addresses else null,
                                note = note,
                                groups = if (groups.isNotEmpty()) groups else null,
                                websites = if (websites.isNotEmpty()) websites else null,
                                events = if (events.isNotEmpty()) events else null,
                                relationships = if (relationships.isNotEmpty()) relationships else null,
                                socialProfiles = if (socialProfiles.isNotEmpty()) socialProfiles else null
                            )
                            contacts.add(contact)
                        }

                        contactsProcessed++
                        if (contactsProcessed % batchSize == 0) {
                            Timber.d("[Mobile] DEBUG [Backup] 已处理 $contactsProcessed/${cursor.count} 个联系人")
                        }

                    } catch (e: Exception) {
                        Timber.e(e, "[Mobile] ERROR [Backup] 处理单个联系人时出错: ${e.message}")
                    }
                }
            } ?: run {
                Timber.e("[Mobile] ERROR [Backup] 备份联系人失败: 无法查询联系人内容提供者")
                return null
            }

            Timber.i("[Mobile] INFO [Backup] 成功读取 ${contacts.size} 个联系人")

        } catch (e: Exception) {
            Timber.e(e, "[Mobile] ERROR [Backup] 备份联系人异常: ${e.message}")
            return null
        }

        return contacts
    }
}
