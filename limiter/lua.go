// the atomic lua script which run in redis
// it prunes old timestamps and checks the count
// and records the new requests at the same time

package limiter

const luaScript = `

    local key = KEYS[1]
		local window = tonumber(ARGV[1])     -- windows in miliseconds 
		local limit = tonumber(ARGV[2])        -- max requests 
		local now  = tonumber(ARGV[3])    -- current time in ms 
		local number = ARGV[4]           -- unique member


		redis.call("ZREMRANGEBYSCORE", key, 0, now - window)
		local count = tonumber(redis.call("ZCARD", key))



		if count < limit then
		    redis.call("ZADD", key, now, member)
				redis.call("PEXPIRE", key, window)
				return {1, limit - count - 1}       -- allowed, remaining 

	  else 
		    local oldest = redis.call("ZRANGE", key, 0, 0, "WITHSCORES")
				local reset = tonumber(oldest[2]) + window

				return {0, 0, reset}                 -- denied, retry after some ms 
		end 
		
 `
