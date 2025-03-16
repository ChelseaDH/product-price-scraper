# 💰 Product Price Scraper

Checking multiple websites to see if your favorite serum or face wash is on sale is faff. Prices change all the time and let's be real, we all have better things so do.

This is where this app comes in! It's a simple scraper that loops through a list of products at the retailers specified and **notifies you when a new lower price is found**.

**How do you get notified?**
- They'll show up in the console
- Or (if you're fancy) they can be sent to a [Matrix](https://matrix.org/docs/chat_basics/matrix-for-im/#what-is-it) chat room

## 🌟 Features
- Scrapes prices from supported retailers at a set interval
- Notifies you only when a **new** lower prices is found (let's avoid the spam!)
- A configurable minimum discount (because who cares about saving £0.05?)
- Matrix integration for notifications

## 🔌 Matrix integration
Want to get those notifications in Matrix as mentioned? Easy! Just set yourself up a bot and configure it in the TOML file ([details below](#matrix-optional)).

## ⚙️ Configuration
The app runs from a single TOML configuration file. Check out the example here: [example.toml](example.toml).

### General settings
- `database` - the name of the app's database
- `interval` - how often scraps should run (e.g. `30m`, `6h`)
- `min_discount` - minimum discount to be notified for _(saves being notified for each tiny price drop - unless you want to)_

### Matrix (optional)
- `home_server` - your Matrix home server URL
- `username` - the bot's username
- `access_token` - the bot's access token
- `room_id` - the ID of the chat room where notifications should go

### Products
This is where you list the products you want to track:

- `name` - name of the product
- `base_price` - the default price to compare against
- `products.links` - a list of retailer URLs for that product

**Supported retailers:**
- [Boots](https://www.boots.com/)
- [Amazon](https://www.amazon.co.uk/)
- [LookFantastic](https://www.lookfantastic.com/)
- [Superdrug](https://www.superdrug.com/)

## 🔨 Build instructions
1. Copy `example.toml` to a file named `config.toml` and insert your desired products and settings 
2. Run the container:
   ```bash
   docker compose up -d
    ```