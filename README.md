# ecommerce-service

A ecommerce service that focuses on product management and order processing.

## ERD

![ERD Diagram](./images/erd.png)

## Services

### **Product Catalog Service**

**Purpose**

* Manages core product information, including descriptions, pricing, images, and categories.
* Provides the source of product data for other services within the e-commerce system.

**Entities**

* **Product**
   * id (integer, primary key)
   * name (string)
   * description (text)
   * price (decimal)
   * sku (string)
   * image_url (string)
   * category_id (integer, foreign key reference to Category)

*  **Category**
    *  id (integer, primary key)
    *  name (string)
    *  parent_category_id (integer, self-referencing foreign key for hierarchies)

**API Endpoints**

* **GET /products/{id}** 
   * Retrieves details for a specific product by its ID.

* **GET /products** 
   * Retrieves a list of products.
   * Supports optional query parameters for:
      * Search terms (search by name, description)
      * Category filtering
      * Pagination 

* **GET /categories**
   * Retrieves a list of available product categories.
   * Potentially supports a hierarchical view (if the category structure allows it).

* **Admin-Level Endpoints (Authentication/Authorization Required)**
   * **POST /products** - Create a new product
   * **PUT /products/{id}** - Update an existing product
   * **DELETE /products/{id}** - Delete a product
   * Similar endpoints for managing categoriesAlright, let's do a similar Markdown style breakdown for the Inventory Management Service.

### **Inventory Management Service**

**Purpose**

* Tracks stock levels for each product.
* Manages reservations/holds when a product is added to a cart.
* Decrements stock upon successful order placement.

**Entities**

* **InventoryItem**
    * id (integer, primary key)
    * product_id (integer, foreign key reference to Product)
    * quantity (integer)
    * available_quantity (integer - potentially derived from quantity and holds)

* **InventoryHold** (Might be optional, depending on how you manage holds)
    * id (integer, primary key)
    * product_id (integer, foreign key reference to Product)
    * quantity (integer)
    * cart_id (integer, foreign key reference to the Cart service) 
    * expiration_time (timestamp)

**API Endpoints**

* **GET /inventory/{product_id}/availability**
    * Retrieves available stock for a specific product

* **PUT /inventory/{product_id}/hold**
    * Places a temporary hold on a specified quantity of a product (associates it with a Cart if relevant).
    * Should include an expiration mechanism for holds.

* **PUT /inventory/{product_id}/decrement**
    * Decrements the stock level for a product, usually triggered after order placement.

*  **Admin-Level Endpoints (Authentication/Authorization Required)**
    * **PUT /inventory/{product_id}/adjust** - Adjust stock levels (add or remove).


## [Project Structure]()



