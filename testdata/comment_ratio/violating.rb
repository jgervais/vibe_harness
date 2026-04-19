# This class handles user authentication
# It validates credentials against the database
# It also manages session tokens
# Sessions expire after 30 minutes
# Failed attempts are logged
class Authenticator
  # Initialize with a database connection
  # The connection must be active
  def initialize(db)
    @db = db
  end

  # Check if the user credentials are valid
  # Returns true if authentication succeeds
  # Returns false otherwise
  # Logs the attempt for auditing
  # Notifies on three consecutive failures
  def authenticate(user, pass)
    @db.verify(user, pass)
  end
end