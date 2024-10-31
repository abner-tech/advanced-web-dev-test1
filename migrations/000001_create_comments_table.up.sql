
--scrpt to the procucts table
CREATE TABLE IF NOT EXISTS products (
    id bigserial PRIMARY KEY,
    name VARCHAR(255), /*name of product*/
    description TEXT, /*detailed Description of product*/
    price DECIMAL(10,2),/*procuct price*/
    category VARCHAR(255),/*Category the product belongs to*/
    image_url VARCHAR(255), /*url to the product image*/
    average_rating DECIMAL(3,2) DEFAULT 0.0, --calculated average rating
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
    version integer NOT NULL DEFAULT 1
);


--script to create the reviews table
CREATE TABLE IF NOT EXISTS reviews (
    id bigserial PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE, --foreign key
    user_name VARCHAR(255), --name of user who posted the rating message and value
    rating INT CHECK(rating BETWEEN 1 AND 5), --rating value between 1 and 5
    review_text TEXT, --rating message
    helpful_count INT DEFAULT 0, --count to see amount of persons who found this rating helpful
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
    version integer NOT NULL DEFAULT 1
);


--function to automatically create average ratings of a product
CREATE OR REPLACE FUNCTION automatic_average_rating()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE products
    SET average_rating = (
        SELECT ROUND(AVG(rating), 2)
        FROM reviews
        WHERE products_id = NEW.product_id
    )
    WHERE id = NEW.product_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


--trigger which excecutes automatic_average_rating() when a new review is added
CREATE OR REPLACE TRIGGER update_product_rating
AFTER INSERT OR UPDATE OR DELETE ON reviews
FOR EACH ROW
EXECUTE FUNCTION automatic_average_rating();