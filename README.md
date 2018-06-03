* [Installation](#installation)
* [How to use](#how-to-use)

<a name="installation"></a>
# Installation #

It requires GO in version `1.10` and `dep` for dependency management.

After fetching sources, run the following command in project's directory:

    dep ensure
    make install # or make build

## Required services ##

It also requires access to Amazon SNS and SendGrid services.

<a name="how-to-use"></a>
# How to use #

## Authorization ##

Application has a very simple authorization method, you need to specify a token by providing it as command line parameter and then use it in http header (`Authorization`).

Run an application with a command:

`emailserv -amazon.key "..," -amazon.secret "..." -sendgrid.key "..." -token "..."
`

## Example request ##

```
{
    "sender": "test@example.com",
    "recipients": [
        "a@example.com",
        "b@example.com"
    ],
    "bcc_recipients": ["c@example.com"],
    "body": "email content"
}
```

## Exmaple response ##

```
{
    "message": "Request tnot valid",
    "validation_errors": [
        {
            field: "sender",
            error: "invalid email address"
        }
    ],
    "error": true
}
```

## NOP client ##

There is a way to debug application without sending emails, in order to do so, run an application with flag: `-nop`. Messages will be logged but not sent.
