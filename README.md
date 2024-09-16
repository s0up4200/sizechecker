# Sizechecker

Companion tool for autobrr for when you want to stop accepting new downloads when a certain disk space limit is met.

## Step 1: Grab the binary

1. Open your terminal.

2. Run the following command to download the binary:

   ```bash
   wget https://github.com/s0up4200/sizechecker/releases/latest/download/sizechecker
   ```

3. Once downloaded, you need to make the file executable. Run the following command:

   ```bash
   chmod +x sizechecker
   ```

### Step 2: Move the Binary to a Directory in Your PATH (optional)

To make the `sizechecker` tool accessible from anywhere, move it to a directory that's part of your system's `PATH`, such as `/usr/local/bin`:

```bash
sudo mv sizechecker /usr/local/bin/
```

Now, you can use the `sizechecker` command from any location in your terminal.

### Step 3: Usage Examples

#### 1. Check Disk Space and Notify via Discord Webhook

The basic usage of the tool is to check the available disk space on a specified directory and optionally send a Discord notification if the available space is below a certain threshold.

**Example:**

```bash
sizechecker --limit=50GB --discord="YOUR_DISCORD_WEBHOOK_URL" /path/to/check
```

- `--limit=50GB`: This sets the minimum required free space in the specified directory. If the available space is less than 50GB, a warning message will be displayed and optionally sent to a Discord webhook.
- `--discord`: This is the Discord webhook URL where the notification will be sent if the disk space is below the specified limit.
- `/path/to/check`: This is the directory where you want to check available space.

#### 2. Set a Cooldown to Avoid Frequent Notifications

You can specify a cooldown period between notifications to prevent the program from sending too many messages in a short time. By default, the cooldown is set to 1 minute if not set.

**Example:**

```bash
sizechecker --discord="YOUR_DISCORD_WEBHOOK_URL" --cooldown=5m --limit=50GB /path/to/check
```

- `--cooldown=5m`: This sets a cooldown period of 5 minutes between notifications. If the disk space check fails within the cooldown period, no additional notification will be sent.
