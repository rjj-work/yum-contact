# 2017.08.19 rjj: Sed file to adapt the Bookshelf code for Yum-Contacts

s!Copyright 2015 Google Inc. All rights reserved.!Adapted from Bookshelf!g

s!library!yum_contacts!g
s!Bookshelf!Contacts!g
s!bookshelf!contacts!g
s!Book!Contact!g
s!book!contact!g

s!Title!FirstName LastName!g
s!Author!Address!g
s!author!address!g
s!Description!Phone!g
s!description!phone!g
#s!Email!Email!g
#s!CreatedByDisplayName!CreatedBy!g
#s!CreateByID!CreatedByID!g
s!PublishedDate!CreatedDate!g
s!publishedDate!createdDate!g

s!Cover Image!Contact Image!g


