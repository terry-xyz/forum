[12/05/2026]
   1. Make home page visible to guests
   2. Proper sessions table
   3. Form validation + escaping/template cleanup
   4. Edit/delete posts and comments
   5. Activity page
   6. Notifications
   7. Image upload
   8. OAuth Google/GitHub
   9. Docker
   10. Tests

---

0. **dont generate same session id again**
   although there is an extremly slight chance of it happening, the random generator might generate the same sessionID twice. we should prevent that.

0. **log errors**
   for example with each "http.Error(...)" we should also send a "fmt.Println("...", err)" to help ourselves as developers but not give too much info to the average user

0. **cookie age**

1. **Use real session tokens instead of user IDs in cookies**  
   In [handlers/auth.go](/home/curadaz/forum/handlers/auth.go:94), the cookie value is `strconv.Itoa(user.ID)`. That is easy to guess. Use a random session ID from `helpers/generator.go` and store/map it server-side.

2. **Hash passwords before storing them**  
   [database/users.go](/home/curadaz/forum/database/users.go:6) stores the raw password. Use `bcrypt` before insert and compare with `bcrypt.CompareHashAndPassword` during login.

3. **Use `html/template` instead of string-building HTML**  
   [handlers/home.go](/home/curadaz/forum/handlers/home.go:53) writes post title/content directly into HTML. That can allow HTML/script injection. Templates escape output automatically.

4. **Move repeated session-cookie parsing into one helper module**  
   `HomeHandler` and `CreatePostHandler` both read the `session` cookie, parse it, and return similar errors. A small helper like `CurrentUser(db, r)` would improve locality and reduce repeated auth logic.

5. **Fix post query efficiency**  
   [handlers/home.go](/home/curadaz/forum/handlers/home.go:44) fetches one user per post. Better: make `GetAllPosts` return posts with author username using a SQL `JOIN`.

6. **Check `rows.Err()` after the loop**  
   In [database/posts.go](/home/curadaz/forum/database/posts.go:38), `rows.Err()` is checked inside the loop. Move it after the loop so scan iteration errors are handled correctly.

7. **Add foreign keys to the schema**  
   [database/schema.sql](/home/curadaz/forum/database/schema.sql:10) has `author_id`, `post_id`, etc., but no `FOREIGN KEY` constraints. Adding them prevents orphan posts/comments/reactions.

8. **Set cookie security options**  
   In [handlers/auth.go](/home/curadaz/forum/handlers/auth.go:94), add `HttpOnly`, `SameSite`, and eventually `Secure` when using HTTPS.

9. **Add input validation before DB writes**  
   `RegisterHandler` and `CreatePostHandler` should reject empty email, username, password, title, or content with `400 Bad Request`.

10. **Close the database in `main`**  
   [main.go](/home/curadaz/forum/main.go:11) opens the DB but never closes it. Add `defer db.Close()` after successful init.
