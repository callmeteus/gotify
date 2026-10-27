package com.github.gotify.client.api;

import com.github.gotify.client.CollectionFormats.*;

import retrofit2.Call;
import retrofit2.http.*;

import okhttp3.RequestBody;
import okhttp3.ResponseBody;
import okhttp3.MultipartBody;

import com.github.gotify.client.model.GotifyInfo;
import com.github.gotify.client.model.Health;
import com.github.gotify.client.model.VersionInfo;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;

public interface InfoApi {
  /**
   * Get health information.
   * 
   * @return Call&lt;Health&gt;
   */
  @GET("health")
  Call<Health> getHealth();
    

  /**
   * Get gotify information.
   * 
   * @return Call&lt;GotifyInfo&gt;
   */
  @GET("gotifyinfo")
  Call<GotifyInfo> getInfo();
    

  /**
   * Get version information.
   * 
   * @return Call&lt;VersionInfo&gt;
   */
  @GET("version")
  Call<VersionInfo> getVersion();
    

}
